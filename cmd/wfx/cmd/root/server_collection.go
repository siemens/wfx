package root

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/coreos/go-systemd/v22/activation"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	genApi "github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/server"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/middleware/plugin"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/spec"
	"golang.org/x/sync/errgroup"
)

type ServerCollection struct {
	once    sync.Once
	cfg     *config.AppConfig
	storage persistence.Storage
	north   *http.Server
	south   *http.Server
	wfx     *api.WfxServer

	pluginMWs    []*plugin.Middleware
	pluginErrors []<-chan error
}

func NewServerCollection(cfg *config.AppConfig, storage persistence.Storage) (*ServerCollection, error) {
	wfx := api.NewWfxServer(storage).
		WithSSEOpts(api.SSEOpts{
			PingInterval:  cfg.SSEPingInterval(),
			GraceInterval: cfg.SSEGraceInterval(),
		})

	swag, _ := genApi.GetSwagger()
	validator := nethttpmiddleware.OapiRequestValidatorWithOptions(swag,
		&nethttpmiddleware.Options{SilenceServersWarning: true})
	corsMW := cors.AllowAll().Handler
	logMW := logging.NewLoggingMiddleware()

	// LIFO
	middlewares := []genApi.MiddlewareFunc{validator, corsMW, logMW}

	pluginMWs := make([]*plugin.Middleware, 0)
	pluginErrors := make([]<-chan error, 0)

	northPluginMWs, err := createPluginMiddlewares(cfg.MgmtPluginsDir())
	if err != nil {
		return nil, fault.Wrap(err)
	}

	for _, mw := range northPluginMWs {
		pluginMWs = append(pluginMWs, mw)
		pluginErrors = append(pluginErrors, mw.Errors())
	}

	northServer, err := createServer(cfg, api.NewNorthboundServer(wfx), middlewares, northPluginMWs)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	southPluginMWs, err := createPluginMiddlewares(cfg.ClientPluginsDir())
	if err != nil {
		return nil, fault.Wrap(err)
	}

	for _, mw := range southPluginMWs {
		pluginMWs = append(pluginMWs, mw)
		pluginErrors = append(pluginErrors, mw.Errors())
	}

	southServer, err := createServer(cfg, api.NewSouthboundServer(wfx), middlewares, southPluginMWs)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	return &ServerCollection{
		cfg:          cfg,
		storage:      storage,
		pluginMWs:    pluginMWs,
		pluginErrors: pluginErrors,
		wfx:          wfx,
		north:        northServer,
		south:        southServer,
	}, nil
}

func (sc *ServerCollection) Start() error {
	sc.wfx.Start()

	cfg := sc.cfg
	schemes := cfg.Schemes()
	// check for socket-based activation; order of sockets: south, north
	systemdListeners, _ := activation.Listeners()
	if len(systemdListeners) > 0 {
		// perform sanity checks
		if len(systemdListeners) != 2 {
			return errors.New("systemd socket-based activation requires two sockets")
		}
		if len(schemes) != 1 || schemes[0] != config.SchemeUnix {
			return errors.New("systemd socket-based activation only supports unix scheme")
		}
	}

	var g errgroup.Group
	for _, scheme := range schemes {
		var northListener, southListener net.Listener
		if len(systemdListeners) > 0 {
			log.Debug().Msg("Using sockets provided by systemd")
			southListener, northListener = systemdListeners[0], systemdListeners[1]
		} else {
			var err error
			northSettings := ListenerSettings{
				Host:    cfg.MgmtHost(),
				Port:    cfg.MgmtPort(),
				TLSHost: cfg.MgmtTLSHost(),
				TLSPort: cfg.MgmtTLSPort(),
				UDSPath: cfg.MgmtUnixSocket(),
			}
			northListener, err = createListener(scheme, northSettings)
			if err != nil {
				return fault.Wrap(err)
			}

			southSettings := ListenerSettings{
				Host:    cfg.ClientHost(),
				Port:    cfg.ClientPort(),
				TLSHost: cfg.ClientTLSHost(),
				TLSPort: cfg.ClientTLSPort(),
				UDSPath: cfg.ClientUnixSocket(),
			}
			southListener, err = createListener(scheme, southSettings)
			if err != nil {
				return fault.Wrap(err)
			}
		}

		isTLS := scheme == config.SchemeHTTPS
		g.Go(func() error {
			defer func() {
				log.Debug().Msg("Northbound goroutine finished")
			}()
			log.Info().
				Bool("tls", isTLS).
				Str("scheme", scheme.String()).
				Str("addr", northListener.Addr().String()).
				Msg("Starting northbound server")

			var err error
			if isTLS {
				err = sc.north.ServeTLS(northListener, cfg.TLSCertificate(), cfg.TLSKey())
			} else {
				err = sc.north.Serve(northListener)
			}
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Err(err).Msg("Northbound server encountered an error")
				return fault.Wrap(err)
			}
			return nil
		})

		g.Go(func() error {
			defer func() {
				log.Debug().Msg("Southbound goroutine finished")
			}()
			log.Info().
				Bool("tls", isTLS).
				Str("scheme", scheme.String()).
				Str("addr", southListener.Addr().String()).
				Msg("Starting southbound server")
			var err error
			if isTLS {
				err = sc.south.ServeTLS(southListener, cfg.TLSCertificate(), cfg.TLSKey())
			} else {
				err = sc.south.Serve(southListener)
			}
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Err(err).Msg("Southbound server encountered an error")
				return fault.Wrap(err)
			}
			return nil
		})
	}

	if len(sc.pluginErrors) > 0 {
		g.Go(func() error {
			defer func() {
				log.Debug().Msg("Plugin reaper finished, triggering shutdown")
				sc.Stop()
			}()

			running := true
			for running {
				for _, chErr := range sc.pluginErrors {
					select {
					case err := <-chErr:
						running = false
						if err != nil {
							log.Err(err).Msg("Received plugin error")
							return err
						}
					default:
						// no errors or channel was closed
					}
				}
				time.Sleep(time.Millisecond * 300)
			}
			return nil
		})
	}

	log.Debug().Msg("Waiting for goroutines to finish")
	err := g.Wait()
	log.Debug().Err(err).Msg("Goroutines finished")

	return fault.Wrap(err)
}

// Stop the server collection and its associated listeners. It's safe to call this method multiple times.
func (sc *ServerCollection) Stop() {
	sc.once.Do(func() {
		timeout := sc.cfg.GracefulTimeout()
		log.Info().Dur("timeout", timeout).Msg("Shutting down server collection")

		// shut down (disconnect) subscribers otherwise we cannot stop the web server due to open connections
		events.ShutdownSubscribers()

		var shutdownGroup sync.WaitGroup
		if sc.north != nil {
			log.Debug().Msg("Shutting down northbound servers")
			shutdownGroup.Add(1)
			go func() {
				defer shutdownGroup.Done()

				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
				defer timeoutCancel()

				_ = sc.north.Shutdown(timeoutCtx)
				log.Debug().Msg("Northbound server shut down complete")
			}()
		}
		if sc.south != nil {
			log.Debug().Msg("Shutting down southbound servers")
			shutdownGroup.Add(1)
			go func() {
				defer shutdownGroup.Done()

				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
				defer timeoutCancel()

				_ = sc.south.Shutdown(timeoutCtx)
				log.Debug().Msg("Southbound server shut down complete")
			}()
		}
		shutdownGroup.Wait()

		log.Debug().Msg("Shutting down plugin middlewares")
		for _, mw := range sc.pluginMWs {
			mw.Stop()
		}

		log.Debug().Msg("Shutting down wfx")
		sc.wfx.Stop()
		log.Debug().Msg("Finished shutting down wfx")

		log.Info().Msg("Server collection shut down complete")
	})
}

func createServer(cfg *config.AppConfig, ssi genApi.StrictServerInterface, baseMWs []genApi.MiddlewareFunc, pluginMWs []*plugin.Middleware) (*http.Server, error) {
	combinedMWs := make([]genApi.MiddlewareFunc, 0, len(baseMWs)+len(pluginMWs))
	combinedMWs = append(combinedMWs, baseMWs...)
	for _, mw := range pluginMWs {
		combinedMWs = append(combinedMWs, mw.Middleware())
	}

	swag, _ := genApi.GetSwagger()
	basePath := errutil.Must(swag.Servers.BasePath())
	strictHandler := genApi.NewStrictHandler(ssi, nil)
	handler := genApi.HandlerWithOptions(strictHandler, genApi.StdHTTPServerOptions{
		BaseURL:     basePath,
		BaseRouter:  createMux(cfg, strictHandler),
		Middlewares: combinedMWs,
	})
	server, err := server.NewHTTPServer(cfg, handler)
	return server, fault.Wrap(err)
}

func createPluginMiddlewares(pluginDir string) ([]*plugin.Middleware, error) {
	plugins, err := loadPlugins(pluginDir)
	if err != nil {
		return nil, fault.Wrap(err)
	}

	pluginMWs := make([]*plugin.Middleware, 0, len(plugins))
	for _, p := range plugins {
		chErr := make(chan error, 1)
		mw, err := plugin.NewMiddleware(p, chErr)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		pluginMWs = append(pluginMWs, mw)
	}
	return pluginMWs, nil
}

type ListenerSettings struct {
	Host    string
	Port    int
	TLSHost string
	TLSPort int
	UDSPath string
}

func createListener(scheme config.Scheme, settings ListenerSettings) (net.Listener, error) {
	var network, addr string
	switch scheme {
	case config.SchemeUnix:
		network = "unix"
		addr = settings.UDSPath
	case config.SchemeHTTP:
		network = "tcp"
		addr = fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	case config.SchemeHTTPS:
		network = "tcp"
		addr = fmt.Sprintf("%s:%d", settings.TLSHost, settings.TLSPort)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", scheme)
	}
	contextLogger := log.With().Str("network", network).Str("addr", addr).Str("scheme", scheme.String()).Logger()
	for attempt := 0; attempt < 30; attempt++ {
		ln, err := net.Listen(network, addr)
		if err == nil {
			contextLogger.Debug().Msg("Created new listener")
			return ln, nil
		}
		contextLogger.Err(err).Msg("Failed to create listener")
		time.Sleep(time.Second)
	}
	return nil, errors.New("failed to create listener")
}

func createMux(cfg *config.AppConfig, server genApi.ServerInterface) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) { server.GetVersion(w, r) })
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { server.GetHealth(w, r) })
	mux.HandleFunc("GET /download/", func(w http.ResponseWriter, r *http.Request) {
		rootDir := cfg.SimpleFileserver()
		enabled := rootDir != ""
		log.Debug().Bool("enabled", enabled).Msg("Received download request")
		if enabled {
			http.StripPrefix("/download", http.FileServer(http.Dir(rootDir))).ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":404,"message":"path /download is not served as simple file server is not enabled"}`))
		}
	})

	for pattern, handler := range spec.Handlers {
		mux.Handle(pattern, handler)
	}

	return mux
}

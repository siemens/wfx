package server

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
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/coreos/go-systemd/v22/activation"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/middleware/plugin"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/spec"
	"github.com/siemens/wfx/ui"
	"golang.org/x/sync/errgroup"
)

var getSpec = sync.OnceValues(api.GetSpec)

type ServerCollection struct {
	once    sync.Once
	cfg     *config.AppConfig
	storage persistence.Storage
	North   *http.Server
	South   *http.Server

	pluginMWs    []*plugin.Middleware
	pluginErrors []<-chan error
}

func NewServerCollection(cfg *config.AppConfig, wfx api.StrictServerInterface, storage persistence.Storage) (*ServerCollection, error) {
	swag, _ := getSpec()
	validator := nethttpmiddleware.OapiRequestValidatorWithOptions(swag,
		&nethttpmiddleware.Options{SilenceServersWarning: true})
	logMW := logging.NewLoggingMiddleware()

	// LIFO
	middlewares := []api.MiddlewareFunc{validator, logMW}

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

	basePath := errutil.Must(swag.Servers.BasePath())
	mux := createMux(cfg, basePath, ui.Enabled)
	northServer, err := createServer(cfg, NewNorthboundServer(wfx), mux, middlewares, northPluginMWs)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	// CORS must wrap the whole server: preflight (OPTIONS) requests match no
	// route of our mux, so a per-route middleware would never see them.
	northHandler := northServer.Handler
	northServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-Client-Id")
		corsOpts := cfg.CORSOpts()
		if !corsOpts.Enabled {
			northHandler.ServeHTTP(w, r)
			return
		}
		cors.New(cors.Options{
			AllowedOrigins:   corsOpts.AllowedOrigins,
			AllowedMethods:   corsOpts.AllowedMethods,
			AllowedHeaders:   corsOpts.AllowedHeaders,
			AllowCredentials: corsOpts.AllowCredentials,
			MaxAge:           corsOpts.MaxAge,
		}).Handler(northHandler).ServeHTTP(w, r)
	})

	southPluginMWs, err := createPluginMiddlewares(cfg.ClientPluginsDir())
	if err != nil {
		return nil, fault.Wrap(err)
	}

	for _, mw := range southPluginMWs {
		pluginMWs = append(pluginMWs, mw)
		pluginErrors = append(pluginErrors, mw.Errors())
	}

	// southbound, UI is always disabled
	mux = createMux(cfg, basePath, false)
	southMiddlewares := append([]api.MiddlewareFunc{}, middlewares...)
	southMiddlewares = append(southMiddlewares, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if clientID := r.Header.Get("X-Client-Id"); clientID != "" {
				r = r.WithContext(persistence.WithClientID(r.Context(), clientID))
			}
			next.ServeHTTP(w, r)
		})
	})
	southServer, err := createServer(cfg, NewSouthboundServer(wfx), mux, southMiddlewares, southPluginMWs)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	southHandler := southServer.Handler
	southServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clientID := r.Header.Get("X-Client-Id"); clientID != "" {
			r = r.WithContext(persistence.WithClientID(r.Context(), clientID))
		}
		southHandler.ServeHTTP(w, r)
	})
	return &ServerCollection{
		cfg:          cfg,
		storage:      storage,
		pluginMWs:    pluginMWs,
		pluginErrors: pluginErrors,
		North:        northServer,
		South:        southServer,
	}, nil
}

func (sc *ServerCollection) Start() error {
	cfg := sc.cfg
	// check for socket-based activation; order of sockets: south, north
	systemdListeners, _ := activation.Listeners()
	if len(systemdListeners) > 0 && len(systemdListeners) != 2 {
		return fault.New("systemd socket-based activation requires two sockets")
	}

	var g errgroup.Group
	start := func(name string, srv *http.Server, ln net.Listener, useTLS bool) {
		log.Info().
			Bool("tls", useTLS).
			Str("addr", ln.Addr().String()).
			Msgf("Starting %s server", name)
		g.Go(func() error {
			defer log.Debug().Msgf("%s goroutine finished", name)
			var err error
			if useTLS {
				err = srv.ServeTLS(ln, cfg.TLSCertificate(), cfg.TLSKey())
			} else {
				err = srv.Serve(ln)
			}
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fault.Wrap(err, fmsg.With(name+" server encountered an error"))
			}
			return nil
		})
	}

	if len(systemdListeners) > 0 {
		log.Debug().Msg("Using sockets provided by systemd")
		start("southbound", sc.South, systemdListeners[0], false)
		start("northbound", sc.North, systemdListeners[1], false)
	} else {
		for _, addr := range cfg.MgmtHosts() {
			ln, err := createListener(addr)
			if err != nil {
				return fault.Wrap(err)
			}
			start("northbound", sc.North, ln, addr.TLS)
		}
		for _, addr := range cfg.ClientHosts() {
			ln, err := createListener(addr)
			if err != nil {
				return fault.Wrap(err)
			}
			start("southbound", sc.South, ln, addr.TLS)
		}
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
							return fault.Wrap(err, fmsg.With("received plugin error"))
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
	log.Debug().Msg("Goroutines finished")

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
		if sc.North != nil {
			log.Debug().Msg("Shutting down northbound servers")
			shutdownGroup.Go(func() {
				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
				defer timeoutCancel()

				_ = sc.North.Shutdown(timeoutCtx)
				log.Debug().Msg("Northbound server shut down complete")
			})
		}
		if sc.South != nil {
			log.Debug().Msg("Shutting down southbound servers")
			shutdownGroup.Go(func() {
				timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
				defer timeoutCancel()

				_ = sc.South.Shutdown(timeoutCtx)
				log.Debug().Msg("Southbound server shut down complete")
			})
		}
		shutdownGroup.Wait()

		log.Debug().Msg("Shutting down plugin middlewares")
		for _, mw := range sc.pluginMWs {
			mw.Stop()
		}

		log.Info().Msg("Server collection shut down complete")
	})
}

func createServer(cfg *config.AppConfig, ssi api.StrictServerInterface, router *http.ServeMux, baseMWs []api.MiddlewareFunc, pluginMWs []*plugin.Middleware) (*http.Server, error) {
	swag, _ := getSpec()
	basePath := errutil.Must(swag.Servers.BasePath())
	strictHandler := api.NewStrictHandlerWithOptions(ssi, nil, api.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			requestLog := logging.LoggerFromCtx(r.Context())
			requestLog.Error().Stack().Err(err).Msg("request failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		},
	})
	router.HandleFunc("GET /version", strictHandler.GetVersion)
	router.HandleFunc("GET /health", strictHandler.GetHealth)
	handler := api.HandlerWithOptions(strictHandler, api.StdHTTPServerOptions{
		BaseURL:     basePath,
		BaseRouter:  router,
		Middlewares: baseMWs,
	})
	for _, mw := range pluginMWs {
		handler = mw.Middleware()(handler)
	}
	server, err := NewHTTPServer(cfg, handler)
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

func createListener(addr config.ListenAddr) (net.Listener, error) {
	contextLogger := log.With().Str("network", addr.Network).Str("addr", addr.Addr).Bool("tls", addr.TLS).Logger()
	ln, err := net.Listen(addr.Network, addr.Addr)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	contextLogger.Debug().Msg("Created new listener")
	return ln, nil
}

func createMux(cfg *config.AppConfig, basePath string, registerUI bool) *http.ServeMux {
	mux := http.NewServeMux()
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

	if registerUI {
		mux.HandleFunc("GET /ui", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui/", http.StatusMovedPermanently)
		})
		oauth := cfg.OAuthOpts()
		uiMux := ui.Mux(basePath, "/ui", ui.OAuthSettings{
			Issuer:   oauth.Issuer,
			ClientID: oauth.ClientID,
			Scope:    oauth.Scope,
		})
		mux.Handle("GET /ui/", http.StripPrefix("/ui", uiMux))
		mux.Handle("GET /favicon.ico", ui.FaviconHandler())
	}

	for pattern, handler := range spec.Handlers {
		mux.Handle(pattern, handler)
	}

	return mux
}

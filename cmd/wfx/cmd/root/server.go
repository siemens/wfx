package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Southclaws/fault"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	genApi "github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/server"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/middleware/plugin"
	"github.com/siemens/wfx/persistence"
)

type ListenerSettings struct {
	Host    string
	Port    int
	TLSHost string
	TLSPort int
	UDSPath string
}

type ServerCollection struct {
	North        *http.Server
	South        *http.Server
	PluginErrors chan error
}

func NewServerCollection(ctx context.Context, cfg *config.AppConfig, storage persistence.Storage, chPluginErrors chan error) (*ServerCollection, error) {
	swag, _ := genApi.GetSwagger()
	validator := nethttpmiddleware.OapiRequestValidatorWithOptions(swag,
		&nethttpmiddleware.Options{SilenceServersWarning: true})
	corsMW := cors.AllowAll().Handler
	logMW := logging.NewLoggingMiddleware()

	// LIFO
	middlewares := []genApi.MiddlewareFunc{validator, corsMW, logMW}
	wfx := api.NewWfxServer(ctx, storage)

	result := ServerCollection{PluginErrors: chPluginErrors}

	northServer, err := createServer(ctx, cfg, api.NewNorthboundServer(wfx), middlewares, cfg.MgmtPluginsDir(), chPluginErrors)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	result.North = northServer

	southServer, err := createServer(ctx, cfg, api.NewSouthboundServer(wfx), middlewares, cfg.ClientPluginsDir(), chPluginErrors)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	result.South = southServer
	return &result, nil
}

func createServer(ctx context.Context, cfg *config.AppConfig, ssi genApi.StrictServerInterface, baseMWs []genApi.MiddlewareFunc, pluginsDir string, chPluginErrors chan error) (*http.Server, error) {
	plugins, err := loadPlugins(pluginsDir, chPluginErrors)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	pluginMWs, err := createPluginMiddlewares(ctx, plugins)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	strictHandler := genApi.NewStrictHandler(ssi, nil)
	combinedMWs := make([]genApi.MiddlewareFunc, 0, len(baseMWs)+len(plugins))
	_ = copy(combinedMWs, baseMWs)
	for _, mw := range pluginMWs {
		combinedMWs = append(combinedMWs, mw)
	}

	swag, _ := genApi.GetSwagger()
	basePath := errutil.Must(swag.Servers.BasePath())
	handler := genApi.HandlerWithOptions(strictHandler, genApi.StdHTTPServerOptions{
		BaseURL:     basePath,
		BaseRouter:  createMux(cfg, strictHandler),
		Middlewares: combinedMWs,
	})
	server, err := server.NewHTTPServer(cfg, handler)
	return server, fault.Wrap(err)
}

func createPluginMiddlewares(ctx context.Context, plugins []plugin.Plugin) ([]func(http.Handler) http.Handler, error) {
	result := make([]func(http.Handler) http.Handler, 0, len(plugins))
	for _, p := range plugins {
		mw, err := plugin.NewMiddleware(ctx, p)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		result = append(result, mw)
	}
	return result, nil
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
	swag, _ := genApi.GetSwagger()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) { server.GetVersion(w, r) })
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { server.GetHealth(w, r) })
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("The requested resource could not be found.\n\nHint: Check /openapi.json to see available endpoints.\n"))
	})
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(swag)
	})
	mux.HandleFunc("GET /download/", func(w http.ResponseWriter, r *http.Request) {
		rootDir := cfg.SimpleFileserver()
		enabled := rootDir != ""
		log.Debug().Bool("enabled", enabled).Msg("Received download request")
		if enabled {
			http.StripPrefix("/download", http.FileServer(http.Dir(rootDir))).ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":404,"message":"path /download was not found"}`))
		}
	})
	return mux
}

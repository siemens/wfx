package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"fmt"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/go-openapi/runtime/flagext"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/api"
	"github.com/siemens/wfx/generated/northbound/restapi"
	"github.com/siemens/wfx/internal/server"
	"github.com/siemens/wfx/middleware"
	"github.com/siemens/wfx/persistence"
)

func createNorthboundServers(storage persistence.Storage) ([]myServer, error) {
	result := make([]myServer, 0, 3)

	var schemes []string
	k.Read(func(k *koanf.Koanf) {
		schemes = k.Strings(schemeFlag)
	})

	for _, scheme := range schemes {
		kind, err := parseServerKind(scheme)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		srv, err := createNorthboundServer(storage, *kind)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		result = append(result, *srv)
	}
	return result, nil
}

func createNorthboundServer(storage persistence.Storage, kind serverKind) (*myServer, error) {
	log.Debug().Str("kind", kind.String()).Msg("Creating northbound server")

	api, err := api.NewNorthboundAPI(storage)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("Failed to create northbound API"))
	}

	swaggerJSON, _ := restapi.SwaggerJSON.MarshalJSON()
	cfg := middleware.Config{
		Config:      k,
		Storage:     storage,
		BasePath:    api.Context().BasePath(),
		SwaggerJSON: swaggerJSON,
	}

	// add our global middlewares
	handler, err := middleware.SetupGlobalMiddleware(cfg, restapi.ConfigureAPI(api))
	if err != nil {
		return nil, fault.Wrap(err)
	}

	var settings server.HTTPSettings
	var mgmtHost, mgmtTLSHost string
	var mgmtPort, mgmtTLSPort int
	k.Read(func(k *koanf.Koanf) {
		settings.MaxHeaderSize = flagext.ByteSize(k.Int(maxHeaderSizeFlag))
		settings.KeepAlive = k.Duration(keepAliveFlag)
		settings.ReadTimeout = k.Duration(readTimeoutFlag)
		settings.WriteTimeout = k.Duration(writeTimoutFlag)
		settings.CleanupTimeout = k.Duration(cleanupTimeoutFlag)

		mgmtHost = k.String(mgmtHostFlag)
		mgmtTLSHost = k.String(mgmtTLSHostFlag)
		mgmtPort = k.Int(mgmtPortFlag)
		mgmtTLSPort = k.Int(mgmtTLSPortFlag)
	})

	switch kind {
	case kindHTTP:
		log.Debug().Msg("Creating http server")
		srv := server.NewHTTPServer(&settings, handler)
		srv.Addr = fmt.Sprintf("%s:%d", mgmtHost, mgmtPort)
		return &myServer{Srv: srv, Kind: kind}, nil
	case kindHTTPS:
		log.Debug().Msg("Creating https server")
		srv := server.NewHTTPServer(&settings, handler)
		srv.Addr = fmt.Sprintf("%s:%d", mgmtTLSHost, mgmtTLSPort)
		var tlsSettings server.TLSSettings
		k.Read(func(k *koanf.Koanf) {
			tlsSettings.TLSCertificate = k.String(tlsCertificateFlag)
			tlsSettings.TLSCertificateKey = k.String(tlsKeyFlag)
			tlsSettings.TLSCACertificate = k.String(tlsCaFlag)
		})
		err := server.ConfigureTLS(srv, &tlsSettings)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		return &myServer{Srv: srv, Kind: kind}, nil
	case kindUnix:
		log.Debug().Msg("Creating unix-domain socket server")
		srv := server.NewHTTPServer(&settings, handler)
		k.Read(func(k *koanf.Koanf) {
			srv.Addr = k.String(mgmtUnixSocketFlag)
		})
		return &myServer{Srv: srv, Kind: kind}, nil
	default:
		return nil, errors.New("unsupported server kind")
	}
}

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
	"github.com/siemens/wfx/generated/southbound/restapi"
	"github.com/siemens/wfx/internal/server"
	"github.com/siemens/wfx/middleware"
	"github.com/siemens/wfx/persistence"
)

func createSouthboundServers(storage persistence.Storage) ([]myServer, error) {
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
		srv, err := createSouthboundServer(storage, *kind)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		result = append(result, *srv)
	}

	return result, nil
}

func createSouthboundServer(storage persistence.Storage, kind serverKind) (*myServer, error) {
	log.Debug().Str("kind", kind.String()).Msg("Creating southbound server")

	api, err := api.NewSouthboundAPI(storage)
	if err != nil {
		return nil, fault.Wrap(err, fmsg.With("Failed to create southbound API"))
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
	var clientHost, clientTLSHost string
	var clientPort, clientTLSPort int
	k.Read(func(k *koanf.Koanf) {
		settings.MaxHeaderSize = flagext.ByteSize(k.Int(maxHeaderSizeFlag))
		settings.KeepAlive = k.Duration(keepAliveFlag)
		settings.ReadTimeout = k.Duration(readTimeoutFlag)
		settings.WriteTimeout = k.Duration(writeTimoutFlag)
		settings.CleanupTimeout = k.Duration(cleanupTimeoutFlag)

		clientHost = k.String(clientHostFlag)
		clientTLSHost = k.String(clientTLSHostFlag)
		clientPort = k.Int(clientPortFlag)
		clientTLSPort = k.Int(clientTLSPortFlag)
	})

	switch kind {
	case kindHTTP:
		log.Debug().Msg("Creating http server")
		srv := server.NewHTTPServer(&settings, handler)
		srv.Addr = fmt.Sprintf("%s:%d", clientHost, clientPort)
		return &myServer{Srv: srv, Kind: kind}, nil
	case kindHTTPS:
		log.Debug().Msg("Creating https server")
		srv := server.NewHTTPServer(&settings, handler)
		srv.Addr = fmt.Sprintf("%s:%d", clientTLSHost, clientTLSPort)
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
			srv.Addr = k.String(clientUnixSocket)
		})
		return &myServer{Srv: srv, Kind: kind}, nil
	default:
		return nil, errors.New("unsupported server kind")
	}
}

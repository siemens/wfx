package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/runtime/flagext"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/internal/server"
)

type myServer struct {
	Srv  *http.Server
	Kind serverKind
}

type serverKind int

const (
	kindHTTP serverKind = iota
	kindHTTPS
	kindUnix
)

func (k serverKind) String() string {
	switch k {
	case kindHTTP:
		return "http"
	case kindHTTPS:
		return "https"
	case kindUnix:
		return "unix"
	}
	panic("unreachable") // non-exhaustive switch statement caught by linter
}

func parseServerKind(scheme string) (*serverKind, error) {
	var result serverKind
	switch scheme {
	case "http":
		result = kindHTTP
	case "https":
		result = kindHTTPS
	case "unix":
		result = kindUnix
	default:
		return nil, fmt.Errorf("unknown scheme: %s", scheme)
	}
	return &result, nil
}

func createServers(schemes []string, handler http.Handler, settings server.HTTPSettings) ([]myServer, error) {
	result := make([]myServer, 0, 3)

	k.Read(func(k *koanf.Koanf) {
		settings.MaxHeaderSize = flagext.ByteSize(k.Int(maxHeaderSizeFlag))
		settings.KeepAlive = k.Duration(keepAliveFlag)
		settings.ReadTimeout = k.Duration(readTimeoutFlag)
		settings.WriteTimeout = k.Duration(writeTimoutFlag)
		settings.CleanupTimeout = k.Duration(cleanupTimeoutFlag)
	})

	for _, scheme := range schemes {
		maybeKind, err := parseServerKind(scheme)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		kind := *maybeKind
		log.Debug().Str("kind", kind.String()).Msg("Creating server")

		switch kind {
		case kindHTTP:
			log.Debug().Msg("Creating http server")
			srv := server.NewHTTPServer(&settings, handler)
			srv.Addr = fmt.Sprintf("%s:%d", settings.Host, settings.Port)
			result = append(result, myServer{Srv: srv, Kind: kind})
		case kindHTTPS:
			log.Debug().Msg("Creating https server")
			srv := server.NewHTTPServer(&settings, handler)
			srv.Addr = fmt.Sprintf("%s:%d", settings.TLSHost, settings.TLSPort)
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
			result = append(result, myServer{Srv: srv, Kind: kind})
		case kindUnix:
			log.Debug().Msg("Creating unix-domain socket server")
			srv := server.NewHTTPServer(&settings, handler)
			srv.Addr = settings.UDSPath
			result = append(result, myServer{Srv: srv, Kind: kind})
		}
	}
	return result, nil
}

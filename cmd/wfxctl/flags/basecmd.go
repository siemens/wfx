package flags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/strfmt"
	"github.com/itchyny/gojq"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/client"
)

type BaseCmd struct {
	EnableTLS bool
	TLSCa     string

	Host    string `validate:"required,hostname_rfc1123"`
	Port    int    `validate:"required"`
	TLSHost string `validate:"required,hostname_rfc1123"`
	TLSPort int    `validate:"required"`
	Socket  string

	MgmtHost    string `validate:"required,hostname_rfc1123"`
	MgmtPort    int    `validate:"required"`
	MgmtTLSHost string `validate:"required,hostname_rfc1123"`
	MgmtTLSPort int    `validate:"required"`
	MgmtSocket  string

	Filter string
	// Strip quotes to make output usable in shell scripts
	RawOutput bool
}

func NewBaseCmd() BaseCmd {
	return BaseCmd{
		EnableTLS: Koanf.Bool(EnableTLSFlag),
		TLSCa:     Koanf.String(TLSCaFlag),

		Host:    Koanf.String(ClientHostFlag),
		Port:    Koanf.Int(ClientPortFlag),
		TLSHost: Koanf.String(ClientTLSHostFlag),
		TLSPort: Koanf.Int(ClientTLSPortFlag),
		Socket:  Koanf.String(ClientUnixSocketFlag),

		MgmtHost:    Koanf.String(MgmtHostFlag),
		MgmtPort:    Koanf.Int(MgmtPortFlag),
		MgmtTLSHost: Koanf.String(MgmtTLSHostFlag),
		MgmtTLSPort: Koanf.Int(MgmtTLSPortFlag),
		MgmtSocket:  Koanf.String(MgmtUnixSocketFlag),

		Filter:    Koanf.String(FilterFlag),
		RawOutput: Koanf.Bool(RawFlag),
	}
}

func (b *BaseCmd) CreateHTTPClient() (*http.Client, error) {
	sockets := make([]string, 0, 2)
	if b.Socket != "" {
		sockets = append(sockets, b.Socket)
	}
	if b.MgmtSocket != "" {
		sockets = append(sockets, b.MgmtSocket)
	}
	if n := len(sockets); n > 0 {
		if n == 2 {
			return nil, fmt.Errorf("you cannot use both --%s and --%s at the same time", ClientUnixSocketFlag, MgmtUnixSocketFlag)
		}
		socket := sockets[0]
		addr, err := net.ResolveUnixAddr("unix", socket)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		return &http.Client{
			Transport: &http.Transport{
				Dial: func(_, _ string) (net.Conn, error) {
					conn, err := net.DialUnix("unix", nil, addr)
					return conn, fault.Wrap(err)
				},
			},
			Timeout: time.Second * 10,
		}, nil
	}

	if !b.EnableTLS {
		return &http.Client{
			Timeout: time.Second * 10,
		}, nil
	}
	caCert, err := os.ReadFile(b.TLSCa)
	if err != nil {
		log.Error().Err(err).Str("tlsCA", b.TLSCa).Msg("Failed to read CA bundle")
		return &http.Client{
			Timeout: time.Second * 10,
		}, nil
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
		Timeout: time.Second * 10,
	}, nil
}

func (b *BaseCmd) CreateClient() *client.WorkflowExecutor {
	var host string
	var schemes []string
	if b.EnableTLS {
		schemes = []string{"https"}
		host = fmt.Sprintf("%s:%d", b.TLSHost, b.TLSPort)
	} else {
		schemes = []string{"http"}
		host = fmt.Sprintf("%s:%d", b.Host, b.Port)
	}

	cfg := client.DefaultTransportConfig().
		WithHost(host).
		WithSchemes(schemes)
	return client.NewHTTPClientWithConfig(strfmt.Default, cfg)
}

func (b *BaseCmd) CreateMgmtClient() *client.WorkflowExecutor {
	var host string
	var schemes []string
	if b.EnableTLS {
		schemes = []string{"https"}
		host = fmt.Sprintf("%s:%d", b.MgmtTLSHost, b.MgmtTLSPort)
	} else {
		schemes = []string{"http"}
		host = fmt.Sprintf("%s:%d", b.MgmtHost, b.MgmtPort)
	}

	cfg := client.DefaultTransportConfig().
		WithHost(host).
		WithSchemes(schemes)
	return client.NewHTTPClientWithConfig(strfmt.Default, cfg)
}

func (b *BaseCmd) DumpResponse(w io.Writer, payload any) error {
	if b.Filter != "" {
		return fault.Wrap(dumpFiltered(payload, b.Filter, b.RawOutput, w))
	}
	return fault.Wrap(dumpPlain(payload, w))
}

func dumpFiltered(payload any, filter string, rawOutput bool, w io.Writer) error {
	query, err := gojq.Parse(filter)
	if err != nil {
		return fault.Wrap(err)
	}

	var input any
	if payloadBytes, ok := payload.([]byte); ok {
		if err := json.Unmarshal(payloadBytes, &input); err != nil {
			return fault.Wrap(err)
		}
	} else {
		data, err := json.Marshal(payload)
		if err != nil {
			return fault.Wrap(err)
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return fault.Wrap(err)
		}
	}

	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fault.Wrap(err)
		}

		if rawOutput {
			if s, ok := v.(string); ok {
				fmt.Fprintf(w, "%s\n", s)
			} else {
				return errors.New("value is not a string. try disabling raw output mode")
			}
		} else {
			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fault.Wrap(err)
			}
			fmt.Fprintln(w, string(b))
		}

	}
	return nil
}

func dumpPlain(payload any, w io.Writer) error {
	if payloadBytes, ok := payload.([]byte); ok {
		fmt.Println(string(payloadBytes))
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return fault.Wrap(enc.Encode(payload))
}

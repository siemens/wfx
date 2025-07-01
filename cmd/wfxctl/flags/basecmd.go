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
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/itchyny/gojq"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/generated/api"
	"github.com/spf13/pflag"
)

const (
	ActorFlag            = "actor"
	ClientHostFlag       = "client-host"
	ClientIDFlag         = "client-id"
	ClientPortFlag       = "client-port"
	ClientTLSHostFlag    = "client-tls-host"
	ClientTLSPortFlag    = "client-tls-port"
	ClientUnixSocketFlag = "client-unix-socket"
	ColorFlag            = "color"
	ConfigFlag           = "config"
	EnableTLSFlag        = "enable-tls"
	FilterFlag           = "filter"
	GroupFlag            = "group"
	HistoryFlag          = "history"
	IDFlag               = "id"
	JobIDFlag            = "job-id"
	LimitFlag            = "limit"
	LogLevelFlag         = "log-level"
	MessageFlag          = "message"
	MgmtHostFlag         = "mgmt-host"
	MgmtPortFlag         = "mgmt-port"
	MgmtTLSHostFlag      = "mgmt-tls-host"
	MgmtTLSPortFlag      = "mgmt-tls-port"
	MgmtUnixSocketFlag   = "mgmt-unix-socket"
	OffsetFlag           = "offset"
	ProgressFlag         = "progress"
	RawFlag              = "raw"
	SortFlag             = "sort"
	StateFlag            = "state"
	TLSCaFlag            = "tls-ca"
	TagFlag              = "tag"
	WorkflowFlag         = "workflow"
	WorkflowNameFlag     = "workflow-name"
	NameFlag             = "name"
	AutoReconnectFlag    = "auto-reconnect"
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
	ColorMode string

	ID        string
	ClientID  string
	ClientIDs []string
	Workflow  string
	Workflows []string
	Tags      []string
	State     string
	Sort      string
	Groups    []string
	Offset    int64
	JobIDs    []string
	History   bool
	Progress  int
	Message   string
	Actor     string
	Name      string
	Limit     int32
}

func NewBaseCmd(f *pflag.FlagSet) BaseCmd {
	k := koanf.New(".")

	if level, err := f.GetString(LogLevelFlag); err == nil {
		if lvl, err := zerolog.ParseLevel(level); err == nil {
			zerolog.SetGlobalLevel(lvl)
		}
	}

	// Load the config files provided in the commandline.
	configFiles, _ := f.GetStringSlice(ConfigFlag)
	log.Debug().Strs("configFiles", configFiles).Msg("Checking config files")
	for _, fname := range configFiles {
		if _, err := os.Stat(fname); err == nil {
			log.Debug().Str("fname", fname).Msg("Loading config file")
			prov := file.Provider(fname)
			if err := k.Load(prov, yaml.Parser()); err != nil {
				log.Fatal().Err(err).Msg("Failed to config file")
			}
		}
	}

	envProvider := env.Provider(".", env.Opt{
		Prefix: "WFX_",
		TransformFunc: func(k string, v string) (string, any) {
			// WFX_LOG_LEVEL becomes log-level
			return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "WFX_")), "_", "-"), nil
		},
	})
	if err := k.Load(envProvider, nil); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Could not load env variables")
	}

	// --log-level becomes log.level
	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to load flags")
	}

	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.Stamp,
	}).With().Timestamp().Logger()
	if lvl, err := zerolog.ParseLevel(k.String(LogLevelFlag)); err == nil {
		zerolog.SetGlobalLevel(lvl)
	}

	return BaseCmd{
		ClientID:    k.String(ClientIDFlag),
		ClientIDs:   k.Strings(ClientIDFlag),
		ColorMode:   k.String(ColorFlag),
		EnableTLS:   k.Bool(EnableTLSFlag),
		Filter:      k.String(FilterFlag),
		Groups:      k.Strings(GroupFlag),
		History:     k.Bool(HistoryFlag),
		Host:        k.String(ClientHostFlag),
		ID:          k.String(IDFlag),
		JobIDs:      k.Strings(JobIDFlag),
		MgmtHost:    k.String(MgmtHostFlag),
		MgmtPort:    k.Int(MgmtPortFlag),
		MgmtSocket:  k.String(MgmtUnixSocketFlag),
		MgmtTLSHost: k.String(MgmtTLSHostFlag),
		MgmtTLSPort: k.Int(MgmtTLSPortFlag),
		Offset:      k.Int64(OffsetFlag),
		Port:        k.Int(ClientPortFlag),
		RawOutput:   k.Bool(RawFlag),
		Socket:      k.String(ClientUnixSocketFlag),
		Sort:        k.String(SortFlag),
		TLSCa:       k.String(TLSCaFlag),
		TLSHost:     k.String(ClientTLSHostFlag),
		TLSPort:     k.Int(ClientTLSPortFlag),
		Tags:        k.Strings(TagFlag),
		Workflow:    k.String(WorkflowFlag),
		Workflows:   k.Strings(WorkflowNameFlag),
		Progress:    k.Int(ProgressFlag),
		Message:     k.String(MessageFlag),
		State:       k.String(StateFlag),
		Actor:       k.String(ActorFlag),
		Name:        k.String(NameFlag),
		Limit:       int32(k.Int(LimitFlag)),
	}
}

func (b *BaseCmd) SortParam() (*api.SortEnum, error) {
	sortRaw := strings.ToLower(b.Sort)
	asc := api.Asc
	desc := api.Desc
	if sortRaw == "" {
		return &asc, nil
	}
	if sortRaw == "asc" {
		return &asc, nil
	}
	if sortRaw == "desc" {
		return &desc, nil
	}
	return nil, fmt.Errorf("invalid sort value: %s", sortRaw)
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
		log.Info().Msg("Using unix-domain socket transport")
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

	tlsConfig := new(tls.Config)
	if b.TLSCa != "" {
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load system cert pool, starting with empty pool")
			caCertPool = x509.NewCertPool()
		}

		log.Debug().Str("tlsCA", b.TLSCa).Msg("Reading CA bundle")
		caCert, err := os.ReadFile(b.TLSCa)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			TLSHandshakeTimeout: time.Second * 10,
		},
		Timeout: time.Second * 10,
	}, nil
}

func (b *BaseCmd) CreateClient() (*api.Client, error) {
	var server string
	swagger := errutil.Must(api.GetSwagger())
	basePath := errutil.Must(swagger.Servers.BasePath())
	if b.EnableTLS {
		server = fmt.Sprintf("https://%s:%d%s", b.TLSHost, b.TLSPort, basePath)
	} else {
		server = fmt.Sprintf("http://%s:%d%s", b.Host, b.Port, basePath)
	}
	log.Debug().Str("server", server).Msg("Creating client")
	httpClient, err := b.CreateHTTPClient()
	if err != nil {
		return nil, fault.Wrap(err)
	}
	client, err := api.NewClient(server, api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return client, nil
}

func (b *BaseCmd) CreateMgmtClient() (*api.Client, error) {
	var server string
	swagger := errutil.Must(api.GetSwagger())
	basePath := errutil.Must(swagger.Servers.BasePath())
	if b.EnableTLS {
		server = fmt.Sprintf("https://%s:%d%s", b.MgmtTLSHost, b.MgmtTLSPort, basePath)
	} else {
		server = fmt.Sprintf("http://%s:%d%s", b.MgmtHost, b.MgmtPort, basePath)
	}
	httpClient, err := b.CreateHTTPClient()
	if err != nil {
		return nil, fault.Wrap(err)
	}
	client, err := api.NewClient(server, api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return client, nil
}

func (b *BaseCmd) ProcessResponse(resp *http.Response, w io.Writer) error {
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	statusCode := resp.StatusCode
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		if err := b.dumpResponse(w, body); err != nil {
			return fault.Wrap(err)
		}
	default:
		var errorResponse api.ErrorResponse
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			errutil.ProcessErrorResponse(w, errorResponse)
		}
		return fmt.Errorf("error: %s", string(body))
	}
	return nil
}

func (b *BaseCmd) dumpResponse(w io.Writer, payload []byte) error {
	if len(payload) == 0 {
		return nil
	}
	if b.Filter != "" {
		return fault.Wrap(dumpFiltered(payload, b.Filter, b.RawOutput, w))
	}
	var body any
	if err := json.Unmarshal(payload, &body); err != nil {
		return fault.Wrap(err)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return fault.Wrap(encoder.Encode(body))
}

func dumpFiltered(payload []byte, filter string, rawOutput bool, w io.Writer) error {
	query, err := gojq.Parse(filter)
	if err != nil {
		return fault.Wrap(err)
	}

	var input any
	if err := json.Unmarshal(payload, &input); err != nil {
		return fault.Wrap(err)
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
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(v); err != nil {
				return fault.Wrap(err)
			}
		}

	}
	return nil
}

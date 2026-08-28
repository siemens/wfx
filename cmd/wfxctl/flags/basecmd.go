package flags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
	"github.com/siemens/wfx/cmd/wfxctl/httpclient"
	"github.com/siemens/wfx/generated/api"
	"github.com/spf13/pflag"
)

const (
	ClientIDFlag         = "client-id"
	ColorFlag            = "color"
	ConfigFlag           = "config"
	CredentialHelperFlag = "credential-helper"
	FilterFlag           = "filter"
	GroupFlag            = "group"
	HeaderFlag           = "header"
	HistoryFlag          = "history"
	HostFlag             = "host"
	IDFlag               = "id"
	JobIDFlag            = "job-id"
	LimitFlag            = "limit"
	LogLevelFlag         = "log-level"
	MessageFlag          = "message"
	OffsetFlag           = "offset"
	ProgressFlag         = "progress"
	RawFlag              = "raw"
	SortFlag             = "sort"
	StateFlag            = "state"
	TLSCaFlag            = "tls-ca"
	TagFlag              = "tag"
	WorkflowFlag         = "workflow"
	NameFlag             = "name"
	AutoReconnectFlag    = "auto-reconnect"
)

type BaseCmd struct {
	TLSCa string

	Host string `validate:"required"`

	Filter string
	// Strip quotes to make output usable in shell scripts
	RawOutput bool
	ColorMode string

	ID               string
	ClientID         string
	ClientIDs        []string
	Workflow         string
	Workflows        []string
	Tags             *[]string
	State            string
	Sort             string
	Groups           []string
	Offset           int64
	JobIDs           []string
	History          bool
	Progress         int
	Message          string
	Name             string
	Limit            int32
	Headers          []string
	CredentialHelper string
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
			log.Debug().Str("fname", fname).Msgf("Loading config file %q", fname)
			prov := file.Provider(fname)
			if err := k.Load(prov, yaml.Parser()); err != nil {
				log.Fatal().Err(err).Msg("Failed to config file")
			}
		}
	}

	envProvider := env.Provider(".", env.Opt{
		Prefix: "WFX",
		TransformFunc: func(k string, v string) (string, any) {
			// WFX_LOG_LEVEL becomes log-level
			key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "WFX_")), "_", "-")
			// Header is read with k.Strings, which does not convert scalar strings.
			if key == HeaderFlag {
				return key, []string{v}
			}
			return key, v
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

	var tags *[]string
	if strs := k.Strings(TagFlag); len(strs) > 0 {
		tags = &strs
	}

	return BaseCmd{
		ClientID:         k.String(ClientIDFlag),
		ClientIDs:        k.Strings(ClientIDFlag),
		ColorMode:        k.String(ColorFlag),
		Filter:           k.String(FilterFlag),
		Groups:           k.Strings(GroupFlag),
		History:          k.Bool(HistoryFlag),
		Host:             k.String(HostFlag),
		ID:               k.String(IDFlag),
		JobIDs:           k.Strings(JobIDFlag),
		Offset:           k.Int64(OffsetFlag),
		RawOutput:        k.Bool(RawFlag),
		Sort:             k.String(SortFlag),
		TLSCa:            k.String(TLSCaFlag),
		Tags:             tags,
		Workflow:         k.String(WorkflowFlag),
		Workflows:        k.Strings(WorkflowFlag),
		Progress:         k.Int(ProgressFlag),
		Message:          k.String(MessageFlag),
		State:            k.String(StateFlag),
		Name:             k.String(NameFlag),
		Limit:            int32(k.Int(LimitFlag)),
		Headers:          k.Strings(HeaderFlag),
		CredentialHelper: k.String(CredentialHelperFlag),
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
	return nil, fault.Newf("invalid sort value: %s", sortRaw)
}

func (b *BaseCmd) CreateHTTPClient() (*http.Client, error) {
	u, err := url.Parse(b.Host)
	if err != nil {
		return nil, fault.New("invalid host URL")
	}
	switch u.Scheme {
	case "http", "https":
		if u.Host == "" {
			u.User = nil
			return nil, fault.Newf("host missing from %q", u.String())
		}
	case "unix":
		if u.Host != "" || u.Path == "" {
			return nil, fault.New("unix host must have form unix:///path/to/socket")
		}
		return &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					conn, err := (&net.Dialer{}).DialContext(ctx, "unix", u.Path)
					return conn, fault.Wrap(err)
				},
			},
			Timeout: time.Second * 10,
		}, nil
	default:
		return nil, fault.Newf("unsupported host scheme %q (want http, https, or unix)", u.Scheme)
	}

	tlsConfig := new(tls.Config)
	if b.TLSCa != "" {
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load system cert pool, starting with empty pool")
			caCertPool = x509.NewCertPool()
		}

		log.Debug().Str("tlsCA", b.TLSCa).Msgf("Reading CA bundle %q", b.TLSCa)
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

func (b *BaseCmd) server() string {
	swagger := errutil.Must(api.GetSpec())
	basePath := errutil.Must(swagger.Servers.BasePath())
	u, err := url.Parse(b.Host)
	if err == nil && u.Scheme == "unix" {
		return "http://localhost" + basePath
	}
	return strings.TrimRight(b.Host, "/") + basePath
}

func (b *BaseCmd) ServerRedacted() string {
	u, err := url.Parse(b.server())
	if err != nil {
		return ""
	}
	u.User = nil
	return u.String()
}

func (b *BaseCmd) CreateClient(opts ...api.ClientOption) (*api.Client, error) {
	server := b.server()
	log.Debug().Msgf("Creating client for %q", b.ServerRedacted())
	httpClient, err := b.CreateHTTPClient()
	if err != nil {
		return nil, fault.Wrap(err)
	}
	editor, err := httpclient.RequestEditor(b.Headers, b.CredentialHelper)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	opts = append([]api.ClientOption{api.WithHTTPClient(httpClient), api.WithRequestEditorFn(editor)}, opts...)
	client, err := api.NewClient(server, opts...)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return client, nil
}

func (b *BaseCmd) CreateClientWithResponses(opts ...api.ClientOption) (*api.ClientWithResponses, error) {
	client, err := b.CreateClient(opts...)
	if err != nil {
		return nil, err
	}
	return &api.ClientWithResponses{ClientInterface: client}, nil
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
		if len(body) == 0 {
			return fault.Newf("empty body, HTTP status %d", statusCode)
		}
		return fault.Newf("error: %s", string(body))
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
				return fault.New("value is not a string. try disabling raw output mode")
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

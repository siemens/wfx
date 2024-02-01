package events

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tmaxmax/go-sse"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	generatedClient "github.com/siemens/wfx/generated/client"
	"github.com/siemens/wfx/generated/client/jobs"
	"github.com/siemens/wfx/generated/model"
)

const (
	jobIDFlag        = "job-id"
	clientIDFlag     = "client-id"
	workflowNameFlag = "workflow-name"
	tagFlag          = "tag"
)

var validator = func(out io.Writer) sse.ResponseValidator {
	return func(r *http.Response) error {
		if r.StatusCode == http.StatusOK {
			return nil
		}

		if r.Body != nil {
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				return fault.Wrap(err)
			}

			errResp := new(model.ErrorResponse)
			if err := json.Unmarshal(b, errResp); err != nil {
				return fault.Wrap(err)
			}
			if len(errResp.Errors) > 0 {
				for _, msg := range errResp.Errors {
					fmt.Fprintf(out, "ERROR: %s (code=%s, logref=%s)\n", msg.Message, msg.Code, msg.Logref)
				}
			}
		}
		return fmt.Errorf("received HTTP status code: %d", r.StatusCode)
	}
}

func init() {
	f := Command.Flags()
	f.StringSlice(jobIDFlag, nil, "job id filter")
	f.StringSlice(clientIDFlag, nil, "client id filter")
	f.StringSlice(workflowNameFlag, nil, "workflow name filter")
	f.StringSlice(tagFlag, nil, "tag filter")
}

type SSETransport struct {
	baseCmd *flags.BaseCmd
	out     io.Writer
}

// Submit implements the runtime.ClientTransport interface.
func (t SSETransport) Submit(op *runtime.ClientOperation) (interface{}, error) {
	cfg := t.baseCmd.CreateTransportConfig()
	rt := client.New(cfg.Host, generatedClient.DefaultBasePath, cfg.Schemes)
	req := errutil.Must(rt.CreateHttpRequest(op))

	httpClient := errutil.Must(t.baseCmd.CreateHTTPClient())
	httpClient.Timeout = 0

	client := sse.Client{
		HTTPClient: httpClient,
		Backoff: sse.Backoff{
			InitialInterval: 5 * time.Second,
			Multiplier:      1.5,
			Jitter:          0.5,
			MaxInterval:     60 * time.Second,
			MaxElapsedTime:  15 * time.Minute,
			MaxRetries:      -1,
		},
		ResponseValidator: validator(t.out),
	}

	conn := client.NewConnection(req)
	unsubscribe := conn.SubscribeMessages(func(event sse.Event) {
		_, _ = os.Stdout.WriteString(event.Data)
		os.Stdout.Write([]byte("\n"))
	})
	defer unsubscribe()

	err := conn.Connect()
	if err != nil && !errors.Is(err, io.EOF) {
		log.Error().Err(err).Msg("Failed to connect to remote server")
		return nil, fault.Wrap(err)
	}

	return jobs.NewGetJobsEventsOK(), nil
}

var Command = &cobra.Command{
	Use:   "events",
	Short: "Subscribe to job events",
	Example: `
wfxctl job events --job-id=1 --job-id=2 --client-id=foo
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		params := jobs.NewGetJobsEventsParams()

		if jobIDs := flags.Koanf.Strings(jobIDFlag); len(jobIDs) > 0 {
			s := strings.Join(jobIDs, ",")
			params.WithJobIds(&s)
		}
		if clientIds := flags.Koanf.Strings(clientIDFlag); len(clientIds) > 0 {
			s := strings.Join(clientIds, ",")
			params.WithClientIds(&s)
		}
		if workflowNames := flags.Koanf.Strings(workflowNameFlag); len(workflowNames) > 0 {
			s := strings.Join(workflowNames, ",")
			params.WithWorkflows(&s)
		}
		if tags := flags.Koanf.Strings(tagFlag); len(tags) > 0 {
			s := strings.Join(tags, ",")
			params.WithTags(&s)
		}

		baseCmd := flags.NewBaseCmd()
		transport := SSETransport{baseCmd: &baseCmd, out: cmd.OutOrStderr()}
		executor := generatedClient.New(transport, strfmt.Default)
		if _, err := executor.Jobs.GetJobsEvents(params); err != nil {
			log.Fatal().Err(err).Msg("Failed to subscribe to job events")
		}
	},
}

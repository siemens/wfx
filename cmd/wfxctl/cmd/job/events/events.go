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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tmaxmax/go-sse"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
)

var validator = func(out io.Writer) sse.ResponseValidator {
	return func(r *http.Response) error {
		fmt.Fprintln(out, "SSE connection established")
		if r.StatusCode == http.StatusOK {
			return nil
		}

		if r.Body != nil {
			defer func() { _ = r.Body.Close() }()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				return fault.Wrap(err)
			}

			errResp := new(api.ErrorResponse)
			if err := json.Unmarshal(b, errResp); err != nil {
				return fault.Wrap(err)
			}
			if errResp.Errors != nil {
				for _, msg := range *errResp.Errors {
					fmt.Fprintf(out, "ERROR: %s (code=%s, logref=%s)\n", msg.Message, msg.Code, msg.Logref)
				}
			}
		}
		return fmt.Errorf("received HTTP status code: %d", r.StatusCode)
	}
}

type SSETransport struct {
	sseClient *sse.Client
	out       io.Writer
}

// Do implements the runtime.ClientTransport interface.
func (t SSETransport) Do(req *http.Request) (*http.Response, error) {
	conn := t.sseClient.NewConnection(req)
	unsubscribe := conn.SubscribeMessages(func(event sse.Event) {
		_, _ = t.out.Write([]byte(event.Data))
		_, _ = t.out.Write([]byte("\n"))
	})
	defer unsubscribe()

	err := conn.Connect()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, fault.Wrap(err)
	}
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Subscribe to job events",
		Example: `
wfxctl job events --job-id=1 --job-id=2 --client-id=foo
`,
		TraverseChildren: true,
		SilenceUsage:     true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f := cmd.Flags()
			baseCmd := flags.NewBaseCmd(f)
			httpClient := errutil.Must(baseCmd.CreateHTTPClient())
			httpClient.Timeout = 0

			// use sane defaults (e.g. auto reconnect) from the default client
			sseClient := sse.DefaultClient
			sseClient.HTTPClient = httpClient
			sseClient.ResponseValidator = validator(cmd.ErrOrStderr())

			autoReconnect, _ := f.GetBool(flags.AutoReconnectFlag)
			if !autoReconnect {
				sseClient.Backoff.MaxRetries = -1
			}

			sseClient.OnRetry = func(_ error, sleep time.Duration) {
				fmt.Fprintf(cmd.ErrOrStderr(), "SSE connection lost. Attempting to reconnect in %v...\n", sleep)
			}
			transport := SSETransport{sseClient: sse.DefaultClient, out: cmd.OutOrStdout()}

			var server string
			swagger := errutil.Must(api.GetSwagger())
			basePath := errutil.Must(swagger.Servers.BasePath())
			if baseCmd.EnableTLS {
				server = fmt.Sprintf("https://%s:%d%s", baseCmd.TLSHost, baseCmd.TLSPort, basePath)
			} else {
				server = fmt.Sprintf("http://%s:%d%s", baseCmd.Host, baseCmd.Port, basePath)
			}

			client, err := api.NewClient(server, api.WithHTTPClient(transport))
			if err != nil {
				return fault.Wrap(err)
			}

			params := new(api.GetJobsEventsParams)
			if jobIDs := baseCmd.JobIDs; len(jobIDs) > 0 {
				s := strings.Join(jobIDs, ",")
				params.JobIds = &s
			}
			if clientIDs := baseCmd.ClientIDs; len(clientIDs) > 0 {
				s := strings.Join(clientIDs, ",")
				params.ClientIDs = &s
			}
			if workflowNames := baseCmd.Workflows; len(workflowNames) > 0 {
				s := strings.Join(workflowNames, ",")
				params.Workflows = &s
			}
			if tags := baseCmd.Tags; len(tags) > 0 {
				s := strings.Join(tags, ",")
				params.Tags = &s
			}
			_, err = client.GetJobsEvents(cmd.Context(), params)
			return fault.Wrap(err)
		},
	}
	f := cmd.Flags()
	f.StringSlice(flags.JobIDFlag, nil, "job id filter")
	f.StringSlice(flags.ClientIDFlag, nil, "client id filter")
	f.StringSlice(flags.WorkflowFlag, nil, "workflow name filter")
	f.StringSlice(flags.TagFlag, nil, "tag filter")
	f.Bool(flags.AutoReconnectFlag, false, "auto reconnect on connection loss")
	return cmd
}

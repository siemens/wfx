package subscribestatus

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
	"os"
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
	idFlag = "id"
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
	f.String(idFlag, "", "job id")
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
		HTTPClient:              httpClient,
		DefaultReconnectionTime: time.Second * 5,
		ResponseValidator:       validator(t.out),
	}

	conn := client.NewConnection(req)
	unsubscribe := conn.SubscribeMessages(func(event sse.Event) {
		_, _ = os.Stdout.Write(event.Data)
		os.Stdout.Write([]byte("\n"))
	})
	defer unsubscribe()

	err := conn.Connect()
	if err != nil {
		return nil, fault.Wrap(err)
	}

	return jobs.NewGetJobsIDStatusSubscribeOK(), nil
}

var Command = &cobra.Command{
	Use:   "subscribe-status",
	Short: "Subscribe to job update events",
	Example: `
wfxctl job subscribe-status --id=1
`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		params := jobs.NewGetJobsIDStatusSubscribeParams().
			WithID(flags.Koanf.String(idFlag))
		if params.ID == "" {
			log.Fatal().Msg("Job ID missing")
		}

		baseCmd := flags.NewBaseCmd()
		transport := SSETransport{baseCmd: &baseCmd, out: cmd.OutOrStderr()}
		executor := generatedClient.New(transport, strfmt.Default)
		if _, err := executor.Jobs.GetJobsIDStatusSubscribe(params); err != nil {
			log.Fatal().Msg("Failed to subscribe to job status")
		}
	},
}

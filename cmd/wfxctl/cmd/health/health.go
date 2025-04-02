package health

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Southclaws/fault"
	"github.com/gookit/color"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/generated/api"
)

type Endpoint struct {
	Name     string
	Server   string
	Response *api.GetHealthResponse
}

const (
	colorNever  = "never"
	colorAlways = "always"
	colorAuto   = "auto"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "health",
		Short:            "Check health of wfx",
		Long:             "Check health wfx",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			b := flags.NewBaseCmd(cmd.Flags())

			var useColor bool
			switch b.ColorMode {
			case colorAlways:
				useColor = true
			case colorAuto:
				useColor = isatty.IsTerminal(os.Stdout.Fd())
			case colorNever:
				useColor = false
			default:
				return fmt.Errorf("unsupported color mode: %s", b.ColorMode)
			}

			swagger := errutil.Must(api.GetSwagger())
			basePath := errutil.Must(swagger.Servers.BasePath())

			allEndpoints := []Endpoint{
				{
					Name:     "northbound",
					Server:   fmt.Sprintf("http://%s:%d%s", b.MgmtHost, b.MgmtPort, basePath),
					Response: &api.GetHealthResponse{Body: []byte("{}")},
				},
				{
					Name:     "southbound",
					Server:   fmt.Sprintf("http://%s:%d%s", b.Host, b.Port, basePath),
					Response: &api.GetHealthResponse{Body: []byte("{}")},
				},
				{
					Name:     "northbound_tls",
					Server:   fmt.Sprintf("https://%s:%d%s", b.MgmtTLSHost, b.MgmtTLSPort, basePath),
					Response: &api.GetHealthResponse{Body: []byte("{}")},
				},
				{
					Name:     "southbound_tls",
					Server:   fmt.Sprintf("https://%s:%d%s", b.TLSHost, b.TLSPort, basePath),
					Response: &api.GetHealthResponse{Body: []byte("{}")},
				},
			}

			httpClient, err := b.CreateHTTPClient()
			if err != nil {
				return fault.Wrap(err)
			}

			var g sync.WaitGroup
			for i, endpoint := range allEndpoints {
				g.Add(1)
				go func() {
					defer g.Done()
					client := errutil.Must(api.NewClientWithResponses(endpoint.Server, api.WithHTTPClient(httpClient)))
					resp, err := client.GetHealthWithResponse(cmd.Context())
					if err != nil {
						log.Warn().Err(err).Msg("Error while checking health")
					} else {
						allEndpoints[i].Response = resp
					}
				}()
			}
			g.Wait()
			prettyReport(cmd.OutOrStdout(), useColor, allEndpoints)
			return nil
		},
	}
	f := cmd.Flags()
	f.String(flags.ColorFlag, colorAuto, fmt.Sprintf("possible values: %s, %s, %s", colorNever, colorAlways, colorAuto))
	return cmd
}

func prettyReport(w io.Writer, useColor bool, allEndpoints []Endpoint) {
	buf := bufio.NewWriter(w)
	defer func() { _ = buf.Flush() }()
	_, _ = buf.WriteString("Health report:\n\n")
	for _, ep := range allEndpoints {
		status := api.Down
		if ep.Response.JSON200 != nil {
			status = ep.Response.JSON200.Status
		} else if ep.Response.JSON503 != nil {
			status = ep.Response.JSON503.Status
		}

		formatter := fmt.Sprint
		if useColor {
			switch status {
			case api.Up:
				formatter = color.FgGreen.Render
			case api.Down:
				formatter = color.FgRed.Render
			default:
				formatter = color.FgYellow.Render
			}
		}
		fmt.Fprintf(buf, "%s\t%s\t(%s)\n", ep.Name, formatter(status), ep.Server)
	}
}

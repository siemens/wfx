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

	"github.com/Southclaws/fault"
	"github.com/gookit/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

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

			endpoint := Endpoint{
				Name:     "wfx",
				Server:   b.ServerRedacted(),
				Response: &api.GetHealthResponse{Body: []byte("{}")},
			}
			client, err := b.CreateClientWithResponses()
			if err != nil {
				return fault.Wrap(err)
			}
			endpoint.Response, err = client.GetHealthWithResponse(cmd.Context())
			prettyReport(cmd.OutOrStdout(), useColor, endpoint)
			if err != nil {
				return fault.Wrap(err)
			}
			if endpoint.Response.JSON200 == nil || endpoint.Response.JSON200.Status != api.Up {
				return fmt.Errorf("wfx is not healthy")
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.String(flags.ColorFlag, colorAuto, fmt.Sprintf("possible values: %s, %s, %s", colorNever, colorAlways, colorAuto))
	return cmd
}

func prettyReport(w io.Writer, useColor bool, ep Endpoint) {
	buf := bufio.NewWriter(w)
	defer func() { _ = buf.Flush() }()
	_, _ = buf.WriteString("Health report:\n\n")
	status := api.Down
	if resp := ep.Response; resp != nil {
		if resp.JSON200 != nil {
			status = resp.JSON200.Status
		} else if resp.JSON503 != nil {
			status = resp.JSON503.Status
		}
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

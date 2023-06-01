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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/gookit/color"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
)

type endpoint struct {
	Name   string                    `json:"name"`
	URL    string                    `json:"url"`
	Status health.AvailabilityStatus `json:"status"`
}

const (
	colorNever  = "never"
	colorAlways = "always"
	colorAuto   = "auto"
	colorFlag   = "color"
)

func init() {
	f := Command.Flags()
	f.String(colorFlag, colorAuto, fmt.Sprintf("possible values: %s, %s, %s", colorNever, colorAlways, colorAuto))
}

var Command = &cobra.Command{
	Use:              "health",
	Short:            "Check health of wfx",
	Long:             "Check health wfx",
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		tty := isatty.IsTerminal(os.Stdout.Fd())

		colorMode := flags.Koanf.String(colorFlag)
		var useColor bool
		switch colorMode {
		case colorAlways:
			useColor = true
		case colorAuto:
			useColor = tty
		case colorNever:
			useColor = false
		default:
			log.Warn().Str("color", colorMode).Msg("Unsupported color mode")
		}

		allEndpoints := []endpoint{
			{
				Name: "northbound",
				URL: fmt.Sprintf("http://%s:%d/health", flags.Koanf.String(flags.MgmtHostFlag),
					flags.Koanf.Int(flags.MgmtPortFlag)),
				Status: health.StatusUnknown,
			},
			{
				Name: "southbound",
				URL: fmt.Sprintf("http://%s:%d/health", flags.Koanf.String(flags.ClientHostFlag),
					flags.Koanf.Int(flags.ClientPortFlag)),
				Status: health.StatusUnknown,
			},
			{
				Name: "northbound_tls",
				URL: fmt.Sprintf("https://%s:%d/health", flags.Koanf.String(flags.MgmtTLSHostFlag),
					flags.Koanf.Int(flags.MgmtTLSPortFlag)),
				Status: health.StatusUnknown,
			},
			{
				Name: "southbound_tls",
				URL: fmt.Sprintf("https://%s:%d/health", flags.Koanf.String(flags.ClientTLSHostFlag),
					flags.Koanf.Int(flags.ClientTLSPortFlag)),
				Status: health.StatusUnknown,
			},
		}

		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				TLSHandshakeTimeout: time.Second * 10,
			},
			Timeout: time.Second * 10,
		}

		for i := range allEndpoints {
			updateStatus(&allEndpoints[i], &client)
		}

		filter, rawOutput := flags.Koanf.String(flags.FilterFlag),
			flags.Koanf.Bool(flags.RawFlag)

		if tty && filter == "" && !rawOutput {
			prettyReport(cmd.OutOrStderr(), useColor, allEndpoints)
		} else {
			baseCmd := flags.NewBaseCmd()
			_ = baseCmd.DumpResponse(cmd.OutOrStdout(), allEndpoints)
		}
	},
}

type simpleHTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

func updateStatus(ep *endpoint, client simpleHTTPClient) {
	resp, err := client.Get(ep.URL)
	if err != nil {
		ep.Status = health.StatusDown
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ep.Status = health.StatusUnknown
		return
	}

	var result health.CheckerResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		ep.Status = health.StatusDown
		return
	}
	ep.Status = health.StatusUp
}

func prettyReport(w io.Writer, useColor bool, allEndpoints []endpoint) {
	buf := bufio.NewWriter(w)
	_, _ = buf.WriteString("Health report:\n\n")
	for _, ep := range allEndpoints {
		f := fmt.Sprint
		if useColor {
			switch ep.Status {
			case health.StatusUp:
				f = color.FgGreen.Render

			case health.StatusDown:
				f = color.FgRed.Render

			default:
				f = color.FgYellow.Render
			}
		}
		fmt.Fprintf(buf, "%s\t%s\t(%s)\n", ep.Name, f(ep.Status), ep.URL)
	}
	buf.Flush()
}

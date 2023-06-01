package loadtest

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/tsenart/vegeta/v12/lib/plot"
)

func dumpResults(metrics *vegeta.Metrics, p *plot.Plot) error {
	log.Debug().Msg("Dumping results")
	resultsDir := fmt.Sprintf("results/%d", time.Now().Unix())
	err := os.MkdirAll(resultsDir, 0o755)
	if err != nil {
		return fault.Wrap(err)
	}

	{
		fmt.Println("*******************************************************************************")
		fmt.Println(" Summary")
		fmt.Println("*******************************************************************************")
		metricsFile, err := os.Create(path.Join(resultsDir, "metrics.txt"))
		if err != nil {
			return fault.Wrap(err)
		}
		defer func() {
			_ = metricsFile.Close()
		}()

		metricsReporter := vegeta.NewTextReporter(metrics)
		if err := metricsReporter.Report(io.MultiWriter(os.Stdout, metricsFile)); err != nil {
			return fault.Wrap(err)
		}
	}

	{
		hdrFile, err := os.Create(path.Join(resultsDir, "histogram.hdr"))
		if err != nil {
			return fault.Wrap(err)
		}
		defer func() {
			_ = hdrFile.Close()
		}()

		hdrReporter := vegeta.NewHDRHistogramPlotReporter(metrics)
		if err := hdrReporter.Report(hdrFile); err != nil {
			return fault.Wrap(err)
		}
	}

	{
		plotFile, err := os.Create(path.Join(resultsDir, "plot.html"))
		if err != nil {
			return fault.Wrap(err)
		}
		defer func() {
			_ = plotFile.Close()
		}()

		if _, err = p.WriteTo(plotFile); err != nil {
			return fault.Wrap(err)
		}
	}

	fmt.Printf("\nResults can be found in directory %s\n", resultsDir)
	return nil
}

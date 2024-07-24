package loadtest

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
	"net/http"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx-loadtest/wfx"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/workflow/dau"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/tsenart/vegeta/v12/lib/plot"
)

const (
	// Threshold of data points above which series are downsampled.
	threshold = 4000

	HostFlag      = "host"
	PortFlag      = "port"
	MgmtHostFlag  = "mgmt-host"
	MgmtPortFlag  = "mgmt-port"
	ReadFreqFlag  = "read-freq"
	WriteFreqFlag = "write-freq"
	DurationFlag  = "duration"
)

var (
	jobCounter uint64
	workflow   = dau.DirectWorkflow()

	queue      = make([]*api.JobStatus, 0, 10)
	queueMutex sync.RWMutex

	host     string
	port     int
	mgmtHost string
	mgmtPort int
)

func Run(k *koanf.Koanf) error {
	host = k.String(HostFlag)
	port = k.Int(PortFlag)
	mgmtHost = k.String(MgmtHostFlag)
	mgmtPort = k.Int(MgmtPortFlag)

	if host == "" || mgmtHost == "" {
		return errors.New("host or mgmtHost not set")
	}

	if err := wfx.CreateWorkflow(mgmtHost, mgmtPort, *workflow); err != nil {
		return fault.Wrap(err)
	}

	duration := k.Duration(DurationFlag)
	writeRate := vegeta.Rate{Freq: k.Int(WriteFreqFlag), Per: time.Second}
	readRate := vegeta.Rate{Freq: k.Int(ReadFreqFlag), Per: time.Second}

	log.Info().
		Str("host", host).
		Int("port", port).
		Str("mgmtHost", mgmtHost).
		Int("mgmtPort", mgmtPort).
		Dur("duration", duration).
		Int("writeRate", writeRate.Freq).
		Int("readRate", readRate.Freq).
		Msg("Starting benchmark")

	var wg sync.WaitGroup

	readerResultChan := make(chan vegeta.Result)
	readerDoneChan := make(chan any)
	wg.Add(1)
	go func() {
		defer wg.Done()
		readTargeter := vegeta.NewStaticTargeter(
			vegeta.Target{
				Method: http.MethodGet,
				URL:    fmt.Sprintf("http://%s:%d/api/wfx/v1/jobs?class=OPEN", host, port),
			},
			vegeta.Target{
				Method: http.MethodGet,
				URL:    fmt.Sprintf("http://%s:%d/api/wfx/v1/workflows", host, port),
			},
		)
		attacker := newAttacker()
		for res := range attacker.Attack(readTargeter, readRate, duration, "Read Jobs") {
			// forward result to reporter
			readerResultChan <- *res
		}
		readerDoneChan <- nil
	}()

	writerResultChan := make(chan vegeta.Result)
	writerDoneChan := make(chan any)
	wg.Add(1)
	go func() {
		defer wg.Done()

		attacker := newAttacker()
		for res := range attacker.Attack(writeTargeter, writeRate, duration, "Generate and update jobs") {
			// forward result to reporter
			writerResultChan <- *res

			if res.Code == http.StatusCreated {
				// we know it must be a job
				var job api.Job
				err := json.Unmarshal(res.Body, &job)
				if err != nil {
					log.Error().Err(err).Bytes("body", res.Body).Msg("Failed to unmarshal body")
					continue
				}

				if job.Status.Context == nil {
					job.Status.Context = &map[string]any{}
				}
				// ensure job id is available
				(*job.Status.Context)["id"] = job.ID

				queueMutex.Lock()
				queue = append(queue, job.Status)
				queueMutex.Unlock()

			} else if res.Code == http.StatusOK {
				// status was updated
				var status api.JobStatus
				err := json.Unmarshal(res.Body, &status)
				if err != nil {
					log.Error().Err(err).Bytes("body", res.Body).Msg("Failed to unmarshal body")
					continue
				}
				// put it back in the queue
				queueMutex.Lock()
				queue = append(queue, &status)
				queueMutex.Unlock()
			}

		}
		writerDoneChan <- nil
	}()

	var metrics vegeta.Metrics
	p := plot.New(
		plot.Title("wfx"),
		plot.Downsample(threshold),
		plot.Label(plot.ErrorLabeler),
	)

	wg.Add(1)
	go func() {
		// collect results
		defer wg.Done()

		doneCounter := 0
		for doneCounter < 2 {
			select {
			case res := <-readerResultChan:
				metrics.Add(&res)
				_ = p.Add(&res)
			case <-readerDoneChan:
				doneCounter++
			case res := <-writerResultChan:
				metrics.Add(&res)
				_ = p.Add(&res)
			case <-writerDoneChan:
				doneCounter++
			}
		}
		metrics.Close()
		p.Close()
	}()

	wg.Wait()
	if err := dumpResults(&metrics, p); err != nil {
		return fault.Wrap(err)
	}
	return nil
}

func newAttacker() *vegeta.Attacker {
	attacker := vegeta.NewAttacker()
	vegeta.Timeout(10 * time.Second)(attacker)
	return attacker
}

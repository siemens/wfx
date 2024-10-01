package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/generated/api"
	flag "github.com/spf13/pflag"
)

var ( // CLI flags
	host         string
	port         int
	clientID     string
	pollInterval time.Duration
)

var (
	ttydCmd *exec.Cmd
	done    atomic.Bool
)

func init() {
	flag.StringVarP(&host, "host", "h", "localhost", "wfx host")
	flag.IntVarP(&port, "port", "p", 8080, "wfx port")
	flag.StringVarP(&clientID, "client-id", "c", "", "client id to use")
	flag.DurationVarP(&pollInterval, "interval", "i", time.Second*10, "polling interval")
}

func main() {
	if _, err := exec.LookPath("ttyd"); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Please install 'ttyd'\n")
		os.Exit(1)
	}

	flag.Parse()
	if clientID == "" {
		fmt.Fprintf(os.Stderr, "FATAL: Argument --client-id is missing\n")
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	log.Println("Press CTRL+C to quit...")

	go worker()

	<-signalChannel
	done.Store(true)
	log.Println("Terminating...")

	terminate()
	log.Println("Bye bye")
}

func worker() {
	workflow := "wfx.workflow.remote.access"
	initialState := "OPEN"
	limit := int32(1)

	swagger := errutil.Must(api.GetSwagger())
	basePath := errutil.Must(swagger.Servers.BasePath())
	server := fmt.Sprintf("http://%s:%d%s", host, port, basePath)
	httpClient := &http.Client{Timeout: time.Second * 10}
	client, err := api.NewClientWithResponses(server, api.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	params := new(api.GetJobsParams)
	params.ParamClientID = &clientID
	params.ParamWorkflow = &workflow
	params.ParamState = &initialState
	params.ParamLimit = &limit

	for !done.Load() {
		log.Println("Polling for new jobs")
		resp, err := client.GetJobsWithResponse(context.Background(), params)
		if err == nil && resp.JSON200 != nil && len(resp.JSON200.Content) > 0 {
			job := &resp.JSON200.Content[0]
			log.Println("Found new job with ID", job.ID)

			args := []string{
				"--writable",
				"--port", "1337",
				"--cwd", "/",
			}
			if credential := job.Definition["credential"].(string); credential != "" {
				args = append(args, "--credential", credential)
			}

			ttydCmd := createTtyCmd(args)

			updateJobStatus(client, job, "OPENING", nil)
			err := ttydCmd.Start()
			if err != nil {
				updateJobStatus(client, job, "FAILED", err)
				continue
			}
			updateJobStatus(client, job, "OPENED", nil)

			if s := job.Definition["timeout"].(string); s != "" {
				if d, err := time.ParseDuration(s); err == nil {
					go func() {
						time.Sleep(d)
						terminate()
					}()
				}
			}
			_ = ttydCmd.Wait()
			updateJobStatus(client, job, "CLOSED", nil)
		} else {
			log.Println("Nothing to do")
			time.Sleep(pollInterval)
		}
	}
}

func updateJobStatus(client *api.ClientWithResponses, job *api.Job, state string, err error) {
	job.Status = &api.JobStatus{State: state}
	if err == nil {
		log.Println("Setting job status to", state)
	} else {
		log.Printf("Setting job status to %s, err=%s\n", state, err)
		job.Status.Context = &map[string]any{
			"error": err.Error(),
		}
	}
	resp, err := client.PutJobsIdStatus(context.Background(), job.ID, nil, *job.Status)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("Failed to update job status:", err)
	}
}

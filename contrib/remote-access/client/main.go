package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/client"
	"github.com/siemens/wfx/generated/client/jobs"
	"github.com/siemens/wfx/generated/model"
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
	queryParams := jobs.NewGetJobsParams().
		WithClientID(&clientID).
		WithWorkflow(&workflow).
		WithState(&initialState).
		WithLimit(&limit)
	cfg := client.DefaultTransportConfig()
	cfg.Host = fmt.Sprintf("%s:%d", host, port)
	c := client.NewHTTPClientWithConfig(strfmt.Default, cfg)

	for !done.Load() {
		log.Println("Polling for new jobs")
		resp, err := c.Jobs.GetJobs(queryParams)
		if err == nil && len(resp.Payload.Content) > 0 && !done.Load() {
			job := resp.Payload.Content[0]
			log.Println("Found new job with ID", job.ID)

			args := []string{
				"--writable",
				"--port", "1337",
				"--cwd", "/",
			}
			if credential := job.Definition["credential"].(string); credential != "" {
				args = append(args, "--credential", credential)
			}
			args = append(args, "bash", "-l")
			ttydCmd = exec.Command("ttyd", args...)
			ttydCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			updateJobStatus(c, job, "OPENING", nil)
			err := ttydCmd.Start()
			if err != nil {
				updateJobStatus(c, job, "FAILED", err)
				continue
			}
			updateJobStatus(c, job, "OPENED", nil)

			if s := job.Definition["timeout"].(string); s != "" {
				if d, err := time.ParseDuration(s); err == nil {
					go func() {
						time.Sleep(d)
						terminate()
					}()
				}
			}
			_ = ttydCmd.Wait()
			updateJobStatus(c, job, "CLOSED", nil)
		} else {
			log.Println("Nothing to do")
			time.Sleep(pollInterval)
		}
	}
}

func updateJobStatus(c *client.WorkflowExecutor, job *model.Job, state string, err error) {
	if err == nil {
		log.Println("Setting job status to", state)
	} else {
		log.Printf("Setting job status to %s, err=%s\n", state, err)
	}
	job.Status = &model.JobStatus{State: state}
	if err != nil {
		job.Status.Context = map[string]any{
			"error": err.Error(),
		}
	}
	_, err = c.Jobs.PutJobsIDStatus(jobs.NewPutJobsIDStatusParams().
		WithID(job.ID).
		WithNewJobStatus(job.Status))
	if err != nil {
		log.Println("Failed to update job status:", err)
	}
}

func terminate() {
	if ttydCmd != nil {
		// ensure ttyd and children are stopped
		pid := ttydCmd.Process.Pid
		ttydCmd = nil
		if err := syscall.Kill(-pid, 0); err == nil {
			// process still running
			log.Println("Sending SIGTERM to", pid)
			syscall.Kill(-pid, syscall.SIGTERM)
		}
		time.Sleep(3 * time.Second)
		if err := syscall.Kill(-pid, 0); err == nil {
			// process still running
			log.Println("Sending SIGKILL to", pid)
			syscall.Kill(-pid, syscall.SIGKILL)
		}
	}
}

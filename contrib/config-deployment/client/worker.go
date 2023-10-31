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
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/client"
	"github.com/siemens/wfx/generated/client/jobs"
	"github.com/siemens/wfx/generated/model"
)

func worker() {
	workflow := "wfx.workflow.config.deployment"
	initialState := "NEW"
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
		log.Println("\n>> Waiting for new job")
		var job *model.Job
		for {
			if done.Load() {
				break
			}
			resp, err := c.Jobs.GetJobs(queryParams)
			if err == nil && len(resp.Payload.Content) > 0 {
				job = resp.Payload.Content[0]
				break
			}
			time.Sleep(pollInterval)
		}
		log.Println("Got new job with ID", job.ID)
		if err := processJob(c, job); err != nil {
			log.Println("Failed to process job:", err)
			updateJobStatus(c, job, "FAILED", nil)
			continue
		}
		log.Println("Successfully processed job")
		updateJobStatus(c, job, "DONE", nil)
	}
}

func processJob(c *client.WorkflowExecutor, job *model.Job) error {
	tmpDir, err := os.MkdirTemp("", "config-deployer")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	updateJobStatus(c, job, "DOWNLOADING", nil)
	url := job.Definition["url"].(string)
	log.Println("Downloading", url)

	tmpFile, err := download(url, tmpDir)
	if err != nil {
		return err
	}

	log.Println("Download successful, starting installation")
	updateJobStatus(c, job, "INSTALLING", nil)

	// run preinstall script
	if preinstall, ok := job.Definition["preinstall"].(string); ok {
		log.Println("Running preinstall script")
		if err := runScript(preinstall); err != nil {
			log.Println("Failed to run preinstall script")
		} else {
			log.Println("Preinstall script was successful")
		}
	}

	// deploy the artifact
	if destination, ok := job.Definition["destination"].(string); ok {
		_ = os.MkdirAll(path.Dir(destination), 0o755)
		src, err := os.Open(tmpFile)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(destination)
		if err != nil {
			return err
		}
		defer dst.Close()

		log.Printf("Copying %s to %s\n", src.Name(), dst.Name())
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	// run postinstall script
	if postinstall, ok := job.Definition["postinstall"].(string); ok {
		log.Println("Running postinstall script")
		if err := runScript(postinstall); err != nil {
			log.Println("Failed to run postinstall script")
		} else {
			log.Println("postinstall script was successful")
		}
	}
	return nil
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

func runScript(script string) error {
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

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
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/generated/api"
)

func worker() {
	workflow := "wfx.workflow.config.deployment"
	initialState := "NEW"
	limit := int32(1)

	params := new(api.GetJobsParams)
	params.ParamClientID = &clientID
	params.ParamWorkflow = &workflow
	params.ParamState = &initialState
	params.ParamLimit = &limit

	swagger := errutil.Must(api.GetSwagger())
	basePath := errutil.Must(swagger.Servers.BasePath())
	server := fmt.Sprintf("http://%s:%d%s", host, port, basePath)
	httpClient := &http.Client{Timeout: time.Second * 10}
	client, err := api.NewClientWithResponses(server, api.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	for !done.Load() {
		log.Println("\n>> Waiting for new job")
		var job *api.Job
		for {
			if done.Load() {
				break
			}
			resp, err := client.GetJobsWithResponse(context.Background(), params)
			if err == nil && resp.JSON200 != nil && len(resp.JSON200.Content) > 0 {
				job = &resp.JSON200.Content[0]
				break
			}
			time.Sleep(pollInterval)
		}
		log.Println("Got new job with ID", job.ID)
		if err := processJob(client, job); err != nil {
			log.Println("Failed to process job:", err)
			updateJobStatus(client, job, "FAILED", nil)
			continue
		}
		log.Println("Successfully processed job")
		updateJobStatus(client, job, "DONE", nil)
	}
}

func processJob(client *api.ClientWithResponses, job *api.Job) error {
	tmpDir, err := os.MkdirTemp("", "config-deployer")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	updateJobStatus(client, job, "DOWNLOADING", nil)
	url := job.Definition["url"].(string)
	log.Println("Downloading", url)

	tmpFile, err := download(url, tmpDir)
	if err != nil {
		return err
	}

	log.Println("Download successful, starting installation")
	updateJobStatus(client, job, "INSTALLING", nil)

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

func runScript(script string) error {
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

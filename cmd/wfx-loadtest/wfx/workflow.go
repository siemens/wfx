package wfx

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
	"net/http"
	"time"

	"github.com/Southclaws/fault"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfxctl/errutil"
	"github.com/siemens/wfx/generated/api"
)

func CreateWorkflow(host string, port int, workflow api.Workflow) error {
	swagger := errutil.Must(api.GetSwagger())
	basePath := errutil.Must(swagger.Servers.BasePath())
	server := fmt.Sprintf("http://%s:%d%s", host, port, basePath)
	log.Info().Str("server", server).Str("name", workflow.Name).Msg("Creating workflow")
	client, err := api.NewClientWithResponses(server, api.WithHTTPClient(&http.Client{
		Timeout: time.Second * 10,
	}))
	if err != nil {
		return fault.Wrap(err)
	}
	{
		resp, err := client.GetWorkflowsNameWithResponse(context.Background(), workflow.Name, nil)
		if err != nil {
			return fault.Wrap(err)
		}
		if resp.JSON200 != nil {
			return nil
		}
	}
	resp, err := client.PostWorkflowsWithResponse(context.Background(), nil, api.PostWorkflowsJSONRequestBody(workflow))
	if err != nil {
		return fault.Wrap(err)
	}
	if resp.JSON201 != nil {
		log.Info().Str("name", workflow.Name).Msg("Created workflow")
		return nil
	}
	body := string(resp.Body)
	log.Warn().Str("body", body).Msg("Failed to create workflow")
	return nil
}

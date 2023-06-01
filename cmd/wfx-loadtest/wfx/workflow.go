package wfx

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/client"
	"github.com/siemens/wfx/generated/client/workflows"
	"github.com/siemens/wfx/generated/model"
)

func CreateWorkflow(host string, port int, workflow *model.Workflow) error {
	log.Debug().Str("name", workflow.Name).Msg("Creating workflow")

	params := workflows.NewPostWorkflowsParams().
		WithHTTPClient(&http.Client{
			Timeout: time.Second * 10,
		}).
		WithWorkflow(workflow)

	cfg := client.DefaultTransportConfig().WithHost(fmt.Sprintf("%s:%d", host, port))
	client := client.NewHTTPClientWithConfig(strfmt.Default, cfg)

	_, err := client.Workflows.PostWorkflows(params)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create workflow")
	}
	log.Info().Str("name", workflow.Name).Msg("Created workflow")
	return nil
}

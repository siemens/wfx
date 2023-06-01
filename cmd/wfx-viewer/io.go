package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/model"
	"gopkg.in/yaml.v3"
)

func readInput(reader io.Reader) *model.Workflow {
	b, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal().Err(err).Msg("Read failed")
	}

	{
		// try JSON
		var workflow model.Workflow
		err := json.Unmarshal(b, &workflow)
		if err == nil {
			return &workflow
		}
	}

	{
		// try YAML
		var workflow model.Workflow
		err := yaml.Unmarshal(b, &workflow)
		if err == nil {
			return &workflow
		}
	}
	log.Fatal().Err(err).Msg("Failed parsing input")
	return nil
}

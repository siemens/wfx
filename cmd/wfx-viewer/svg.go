package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/model"
)

func generateSvg(out io.Writer, krokiURL string, workflow *model.Workflow) error {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	w, err := writer.CreateFormFile("file", "workflow")
	if err != nil {
		return fault.Wrap(err)
	}

	generatePlantUML(w, workflow)

	err = writer.Close()
	if err != nil {
		return fault.Wrap(err)
	}

	req, err := http.NewRequest(http.MethodPost, krokiURL, body)
	if err != nil {
		return fault.Wrap(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fault.Wrap(err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fault.Wrap(err)
	}
	if err := resp.Body.Close(); err != nil {
		return fault.Wrap(err)
	}

	return nil
}

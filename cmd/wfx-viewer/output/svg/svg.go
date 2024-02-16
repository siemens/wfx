package svg

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
	"github.com/siemens/wfx/cmd/wfx-viewer/output/plantuml"
	"github.com/siemens/wfx/generated/model"
	"github.com/spf13/pflag"
)

const krokiURLFlag = "kroki-url"

type Generator struct {
	f *pflag.FlagSet
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (s *Generator) RegisterFlags(f *pflag.FlagSet) {
	f.String(krokiURLFlag, "https://kroki.io/plantuml/svg", "url to kroki (used for svg)")
	s.f = f
}

func (s *Generator) Generate(out io.Writer, workflow *model.Workflow) error {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	w, err := writer.CreateFormFile("file", "workflow")
	if err != nil {
		return fault.Wrap(err)
	}

	if err := plantuml.NewGenerator().Generate(w, workflow); err != nil {
		return fault.Wrap(err)
	}

	err = writer.Close()
	if err != nil {
		return fault.Wrap(err)
	}

	krokiURL, err := s.f.GetString(krokiURLFlag)
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

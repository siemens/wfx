package httpclient

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/api"
	"golang.org/x/net/http/httpguts"
)

// parseHeader parses a single "Name: Value" pair as accepted by curl's -H flag.
func parseHeader(header string) (string, string, error) {
	name, value, found := strings.Cut(header, ":")
	if !found || !httpguts.ValidHeaderFieldName(name) || !httpguts.ValidHeaderFieldValue(value) {
		return "", "", errors.New("invalid header, expected format 'Name: Value'")
	}
	return name, strings.Trim(value, " \t"), nil
}

// RequestEditor returns a RequestEditorFn which adds custom headers to each
// outgoing request. Headers are parsed eagerly so invalid input fails immediately.
func RequestEditor(values []string) (api.RequestEditorFn, error) {
	headers := make(http.Header, len(values))
	for _, header := range values {
		name, value, err := parseHeader(header)
		if err != nil {
			return nil, fault.Wrap(err)
		}
		headers.Add(name, value)
	}
	return func(_ context.Context, req *http.Request) error {
		for name, values := range headers {
			if name == "Host" {
				req.Host = values[len(values)-1]
				continue
			}
			req.Header[name] = values
		}
		return nil
	}, nil
}

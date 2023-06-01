package logging

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
	"net/http"

	"github.com/Southclaws/fault"
)

type requestReader struct {
	requestBody    *bytes.Buffer
	originalReader io.ReadCloser
	teeReader      *io.Reader
}

func newMyRequestReader(r *http.Request) requestReader {
	var buf bytes.Buffer
	tee := io.TeeReader(r.Body, &buf)
	myReader := requestReader{requestBody: &buf, originalReader: r.Body, teeReader: &tee}
	r.Body = myReader
	return myReader
}

// Read reads up to len(p) bytes into p.
func (r requestReader) Read(p []byte) (int, error) {
	n, err := (*r.teeReader).Read(p)
	return n, fault.Wrap(err)
}

func (r requestReader) Close() error {
	return fault.Wrap(r.originalReader.Close())
}

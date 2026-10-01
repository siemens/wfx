// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0

package gron

import (
	"bufio"
	"bytes"
	"errors"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Southclaws/fault"
)

// Middleware returns gron when request explicitly accepts application/gron.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addVary(w.Header(), "Accept")
		if !acceptsGron(r.Header.Values("Accept")) {
			next.ServeHTTP(w, r)
			return
		}

		writer := &responseWriter{ResponseWriter: w, wantsGron: true}
		next.ServeHTTP(writer, r)
		writer.finish()
	})
}

func acceptsGron(values []string) bool {
	for _, value := range values {
		for _, mediaRange := range splitAccept(value) {
			mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(mediaRange))
			if err != nil || !strings.EqualFold(mediaType, "application/gron") {
				continue
			}
			quality := 1.0
			if q, ok := params["q"]; ok {
				quality, err = strconv.ParseFloat(q, 64)
				if err != nil || !(quality > 0 && quality <= 1) {
					continue
				}
			}
			return quality > 0
		}
	}
	return false
}

func splitAccept(value string) []string {
	var parts []string
	start, quoted, escaped := 0, false, false
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\\':
			if quoted {
				escaped = !escaped
			}
		case '"':
			if !escaped {
				quoted = !quoted
			}
			escaped = false
		case ',':
			if !quoted {
				parts = append(parts, value[start:i])
				start = i + 1
			}
			escaped = false
		default:
			escaped = false
		}
	}
	return append(parts, value[start:])
}

func addVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, name := range strings.Split(existing, ",") {
			if name == "*" || strings.EqualFold(strings.TrimSpace(name), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}

type responseWriter struct {
	http.ResponseWriter
	body        bytes.Buffer
	statusCode  int
	passthrough bool
	wantsGron   bool
	hijacked    bool
	flushErr    error
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.statusCode == 0 && !w.hijacked {
		w.statusCode = statusCode
	}
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.flushErr != nil {
		return 0, w.flushErr
	}
	if w.passthrough {
		n, err := w.ResponseWriter.Write(body)
		return n, fault.Wrap(err)
	}
	n, err := w.body.Write(body)
	return n, fault.Wrap(err)
}

func (w *responseWriter) Flush() {
	if w.hijacked {
		return
	}
	w.Header().Del("Content-Length")
	if !w.passthrough {
		w.passthrough = true
		w.commit()
		_, w.flushErr = w.body.WriteTo(w.ResponseWriter)
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("http.Hijacker interface not supported")
	}
	conn, rw, err := hijacker.Hijack()
	if err == nil {
		w.hijacked = true
	}
	return conn, rw, fault.Wrap(err)
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) finish() {
	if w.hijacked {
		return
	}
	contentType, _, _ := mime.ParseMediaType(w.Header().Get("Content-Type"))
	if w.wantsGron && !w.passthrough && strings.EqualFold(contentType, "application/json") {
		var encoded bytes.Buffer
		if err := Encode(&encoded, bytes.NewReader(w.body.Bytes())); err == nil {
			addVary(w.Header(), "Accept")
			w.Header().Set("Content-Type", "application/gron")
			w.Header().Del("Content-Length")
			w.body = encoded
		}
	}
	w.commit()
	_, _ = w.body.WriteTo(w.ResponseWriter)
}

func (w *responseWriter) commit() {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(w.statusCode)
}

// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0

package gron

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	const input = `{"name":"wfx","items":[1,{}],"odd-key":null}`
	const want = "json = {};\njson.items = [];\njson.items[0] = 1;\njson.items[1] = {};\njson.name = \"wfx\";\njson[\"odd-key\"] = null;\n"

	var got bytes.Buffer
	if err := Encode(&got, strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	if got.String() != want {
		t.Fatalf("Encode() = %q, want %q", got.String(), want)
	}
}

func TestMiddlewareNegotiatesGron(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept", `application/json, application/gron; q=0.8`)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got, want := recorder.Header().Get("Content-Type"), "application/gron"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	if got, want := recorder.Header().Get("Vary"), "Accept"; got != want {
		t.Fatalf("Vary = %q, want %q", got, want)
	}
	if got, want := recorder.Body.String(), "json = {};\njson.ok = true;\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestMiddlewareRejectsZeroQuality(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept", "application/gron;q=0")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if got, want := recorder.Body.String(), `{"ok":true}`; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
	if got, want := recorder.Header().Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	if got := recorder.Header().Get("Vary"); got != "Accept" {
		t.Fatalf("Vary = %q, want Accept", got)
	}
}

func TestMiddlewarePreservesSSE(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		_, _ = io.WriteString(w, "data: event\n\n")
		flusher.Flush()
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Accept", "application/gron")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", got)
	}
	if strings.Contains(recorder.Body.String(), "json =") {
		t.Fatalf("SSE body was converted: %q", recorder.Body.String())
	}
}

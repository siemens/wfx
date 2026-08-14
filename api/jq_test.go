package api

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisitMethod(t *testing.T) {
	obj := map[string]string{"foo": "bar"}
	filter := `.foo`
	jqFilter := NewJQFilter(t.Context(), filter, obj, config.JQOpts{FilterTimeout: time.Minute})

	// Get the type and value of the struct
	structType := reflect.TypeOf(jqFilter)
	structValue := reflect.ValueOf(jqFilter)

	// Iterate through the methods of the struct
	for i := 0; i < structType.NumMethod(); i++ {
		method := structType.Method(i)
		if strings.HasPrefix(method.Name, "Visit") {
			t.Run(method.Name, func(t *testing.T) {
				// Get the method by name
				methodValue := structValue.MethodByName(method.Name)

				// Call the method
				recorder := httptest.NewRecorder()
				args := []reflect.Value{reflect.ValueOf(recorder)}
				_ = methodValue.Call(args)
				resp := recorder.Result()
				assert.Equal(t, filter, resp.Header.Get("X-Response-Filter"))
				body, _ := io.ReadAll(resp.Body)
				assert.Equal(t, "\"bar\"\n", string(body))
			})
		}
	}
}

func TestApplyFilterInvalid(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := applyFilter(t.Context(), recorder, nil, "invalid filter", config.JQOpts{FilterTimeout: time.Minute})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "wfx.invalidResponseFilter")
}

type NoMarshal struct{}

func (NoMarshal) MarshalJSON() ([]byte, error) {
	return nil, errors.New("no marshal")
}

func TestApplyFilterMarshalError(t *testing.T) {
	err := applyFilter(t.Context(), nil, NoMarshal{}, ".", config.JQOpts{FilterTimeout: time.Minute})
	assert.Error(t, err)
}

func TestApplyFilterRuntimeError(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := applyFilter(t.Context(), recorder, nil, `1, error("boom")`, config.JQOpts{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	// no part of the successful output must leak into the error response
	assert.NotContains(t, recorder.Body.String(), "1\n")
	assert.Empty(t, recorder.Header().Get("X-Response-Filter"))
}

func TestApplyFilterResultTooLarge(t *testing.T) {
	recorder := httptest.NewRecorder()

	maxResponseSize := 1024
	err := applyFilter(t.Context(), recorder, nil, "range(0; 1000000000)", config.JQOpts{FilterMaxResponseSize: maxResponseSize})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "wfx.invalidResponseFilter")
	assert.Less(t, recorder.Body.Len(), maxResponseSize)
}

func TestLimitedBufferRejectsOversizedWrite(t *testing.T) {
	buf := limitedBuffer{max: 4}

	n, err := buf.Write([]byte("12345"))

	assert.Zero(t, n)
	assert.ErrorIs(t, err, errFilterResultTooLarge)
	assert.Zero(t, buf.Len())
}

func TestApplyFilterUsesRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	recorder := httptest.NewRecorder()

	err := applyFilter(ctx, recorder, map[string]string{"foo": "bar"}, ".foo", config.JQOpts{FilterTimeout: time.Minute})

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, recorder.Body.String())
}

func TestApplyFilterCancelsLongRunningFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	recorder := httptest.NewRecorder()
	started := time.Now()

	err := applyFilter(ctx, recorder, nil, "repeat(0)", config.JQOpts{FilterTimeout: time.Minute})

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(started), time.Second)
}

func TestApplyFilterTimeout(t *testing.T) {
	recorder := httptest.NewRecorder()
	started := time.Now()

	err := applyFilter(t.Context(), recorder, nil, "repeat(0)", config.JQOpts{FilterTimeout: 10 * time.Millisecond})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), context.DeadlineExceeded.Error())
	assert.Less(t, time.Since(started), time.Second)
}

func TestApplyFilterNoTimeout(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := applyFilter(t.Context(), recorder, map[string]string{"foo": "bar"}, ".foo", config.JQOpts{})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "\"bar\"\n", recorder.Body.String())
}

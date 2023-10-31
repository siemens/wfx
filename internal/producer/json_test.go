package producer

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONProducer_NilSliceEmpty(t *testing.T) {
	var nilSlice []string
	buf := new(bytes.Buffer)
	err := JSONProducer().Produce(buf, nilSlice)
	require.NoError(t, err)
	assert.JSONEq(t, "null", buf.String())
}

func TestJSONProducer_NilMapEmpty(t *testing.T) {
	var nilMap map[string]string
	buf := new(bytes.Buffer)
	err := JSONProducer().Produce(buf, nilMap)
	require.NoError(t, err)
	assert.JSONEq(t, "null", buf.String())
}

type TestResponseWriter struct {
	Headers map[string][]string
	Output  []byte
}

func (w *TestResponseWriter) Header() http.Header {
	return w.Headers
}

func (w *TestResponseWriter) Write(buf []byte) (int, error) {
	w.Output = buf
	return len(buf), nil
}

func (w *TestResponseWriter) WriteHeader(int) {}

func TestJSONProducer_ResponseFilter(t *testing.T) {
	headers := make(map[string][]string)
	headers["X-Response-Filter"] = []string{".name"}

	writer := TestResponseWriter{
		Headers: headers,
		Output:  make([]byte, 2048),
	}

	data := make(map[string]string)
	data["name"] = "foo"
	data["id"] = "42"

	prod := JSONProducer()
	err := prod.Produce(&writer, data)
	require.NoError(t, err)
	assert.Equal(t, `"foo"`, string(writer.Output))
}

type BadMarshal struct{}

func (bm *BadMarshal) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Marshaling this struct always fails")
}

func TestJSONProducer_BadMarshal(t *testing.T) {
	badMarshal := BadMarshal{}
	buf := new(bytes.Buffer)
	err := JSONProducer().Produce(buf, &badMarshal)
	assert.NotNil(t, err)
}

func TestJSONProducer_BadFilter(t *testing.T) {
	headers := make(map[string][]string)
	headers["X-Response-Filter"] = []string{"!!!"}
	writer := TestResponseWriter{
		Headers: headers,
		Output:  make([]byte, 2048),
	}

	s := "hello world"
	err := JSONProducer().Produce(&writer, &s)
	assert.NotNil(t, err)
}

type BadWriter struct {
	Headers map[string][]string
}

func (w *BadWriter) Header() http.Header {
	return w.Headers
}

func (w *BadWriter) Write([]byte) (int, error) {
	return 0, errors.New("This writer always fails")
}

func (w *BadWriter) WriteHeader(int) {}

func TestJSONProducer_BadWriter(t *testing.T) {
	headers := make(map[string][]string)
	headers["X-Response-Filter"] = []string{".hello"}
	writer := BadWriter{Headers: headers}

	data := make(map[string]string)
	data["hello"] = "world"
	err := JSONProducer().Produce(&writer, &data)
	assert.NotNil(t, err)
}

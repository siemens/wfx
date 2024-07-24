package api

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitMethod(t *testing.T) {
	obj := map[string]string{"foo": "bar"}
	filter := `.foo`
	jqFilter := NewJQFilter(filter, obj)

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
	err := applyFilter(nil, nil, "invalid filter")
	assert.Error(t, err)
}

type NoMarshal struct{}

func (NoMarshal) MarshalJSON() ([]byte, error) {
	return nil, errors.New("no marshal")
}

func TestApplyFilterMarshalError(t *testing.T) {
	err := applyFilter(nil, NoMarshal{}, "")
	assert.Error(t, err)
}

//go:build testing

package persistence

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	mock "github.com/stretchr/testify/mock"
)

func NewHealthyMockStorage(t *testing.T) *MockStorage {
	m := new(MockStorage)
	m.Mock.Test(t)
	m.On("CheckHealth", mock.Anything).Return(nil)
	return m
}

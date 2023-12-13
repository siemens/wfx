package persistence

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageAPI(t *testing.T) {
	storage1 := NewMockStorage(t)

	RegisterStorage("storage1", storage1)
	actual := GetStorage("storage1")
	assert.Same(t, storage1, actual)

	storage2 := NewMockStorage(t)
	RegisterStorage("storage2", storage2)
	all := Storages()
	assert.Len(t, all, 2)
	assert.Nil(t, GetStorage("foo"))
}

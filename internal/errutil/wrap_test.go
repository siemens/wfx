package errutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap2(t *testing.T) {
	t.Parallel()

	t.Run("no error", func(t *testing.T) {
		t.Parallel()

		val, err := Wrap2("foo", nil)
		assert.Equal(t, "foo", val)
		assert.Nil(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		val, err := Wrap2[any](nil, errors.New("test"))
		assert.Nil(t, val)
		assert.NotNil(t, err)
	})

	t.Run("both", func(t *testing.T) {
		t.Parallel()

		val, err := Wrap2("foo", errors.New("test"))
		assert.Equal(t, "foo", val)
		assert.NotNil(t, err)
	})
}

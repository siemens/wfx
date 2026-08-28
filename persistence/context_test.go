package persistence

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientIDContext(t *testing.T) {
	clientID, ok := ClientIDFromContext(context.Background())
	assert.False(t, ok)
	assert.Empty(t, clientID)

	ctx := WithClientID(context.Background(), "client-1")
	clientID, ok = ClientIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "client-1", clientID)
}

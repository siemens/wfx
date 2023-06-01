package version

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/siemens/wfx/cmd/wfxctl/flags"
	"github.com/siemens/wfx/middleware/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	server := httptest.NewServer(version.NewVersionMiddleware(nil))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	_ = flags.Koanf.Set(flags.ClientHostFlag, parsedURL.Hostname())
	_ = flags.Koanf.Set(flags.ClientPortFlag, parsedURL.Port())

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	err := Command.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "version:")
}

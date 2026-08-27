package root

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"os"
	"testing"

	"go.uber.org/goleak"
)

const testHost = "localhost"

func TestMain(m *testing.M) {
	_ = os.Setenv("WFX_CLIENT_HOST", testHost)
	_ = os.Setenv("WFX_CLIENT_TLS_HOST", testHost)
	goleak.VerifyTestMain(m)
}

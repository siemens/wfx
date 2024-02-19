package mermaid

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	approvals.UseFolder("testdata")
	goleak.VerifyTestMain(m)
}

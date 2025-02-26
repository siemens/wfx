//go:build no_plugin

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

package root

func init() {
	buildTags = append(buildTags, "no_plugin")
}

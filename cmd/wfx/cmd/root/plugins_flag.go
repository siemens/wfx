package root

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

const (
	clientPluginsDirFlag = "client-plugins-dir"
	mgmtPluginsDirFlag   = "mgmt-plugins-dir"
)

func init() {
	f := Command.PersistentFlags()

	_ = Command.MarkPersistentFlagDirname(clientPluginsDirFlag)
	f.String(clientPluginsDirFlag, "", "directory containing client plugins")

	_ = Command.MarkPersistentFlagDirname(mgmtPluginsDirFlag)
	f.String(mgmtPluginsDirFlag, "", "directory containing management plugins")
}

package flags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"github.com/knadh/koanf/v2"
)

const (
	ConfigFlag = "config"

	ClientHostFlag       = "client-host"
	ClientPortFlag       = "client-port"
	ClientUnixSocketFlag = "client-unix-socket"
	ClientTLSHostFlag    = "client-tls-host"
	ClientTLSPortFlag    = "client-tls-port"

	MgmtHostFlag       = "mgmt-host"
	MgmtPortFlag       = "mgmt-port"
	MgmtUnixSocketFlag = "mgmt-unix-socket"
	MgmtTLSHostFlag    = "mgmt-tls-host"
	MgmtTLSPortFlag    = "mgmt-tls-port"

	TLSCaFlag = "tls-ca"

	FilterFlag    = "filter"
	RawFlag       = "raw"
	EnableTLSFlag = "enable-tls"
)

var Koanf = koanf.New(".")

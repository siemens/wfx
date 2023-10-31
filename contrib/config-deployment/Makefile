# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
.PHONY: build
build: config-deployer

.PHONY: config-deployer
config-deployer:
	@CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o $@ ./client
	@echo "$@: build successful"

.PHONY: clean
clean:
	$(RM) config-deployer

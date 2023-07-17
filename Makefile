# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --jobs=$(shell nproc)

DESTDIR ?=
prefix ?= /usr/local

GO_TAGS = sqlite,postgres,mysql

export CGO_ENABLED=1

LD_FLAGS := -s -w
STATIC ?= 0
ifeq ($(STATIC), 1)
LD_FLAGS := -linkmode external -extldflags "-static" -s -w
endif

ALL_TARGETS := wfx wfxctl wfx-viewer wfx-loadtest

.PHONY: default
default:
	@make -s $(ALL_TARGETS)

.PHONY: test
test:
	go test -race -coverprofile=coverage.out -covermode=atomic -timeout 30s ./... "--tags=sqlite,testing"

.PHONY: install
install:
	for target in $(ALL_TARGETS); do \
		install -m 0755 -D $$target $(DESTDIR)$(prefix)/bin/$$target ; \
	done

.PHONY: wfx wfxctl wfx-loadtest wfx-viewer
wfx wfxctl wfx-loadtest wfx-viewer:
	@echo "Building $@"
	@go build -trimpath -tags=$(GO_TAGS) \
		-ldflags '$(LD_FLAGS) -X github.com/siemens/wfx/cmd/$@/metadata.Commit=$(shell git rev-parse HEAD | tr -d [:space:]) -X github.com/siemens/wfx/cmd/$@/metadata.Date=$(shell date -Iseconds)' \
		./cmd/$@

.PHONY: clean
clean:
	@$(RM) $(ALL_TARGETS) *.exe
	@find . \( -name "*.db" -o -name "*.db-wal" -o -name "*.db-shm" -o -name "*.out" \) -delete

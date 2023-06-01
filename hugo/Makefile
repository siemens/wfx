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

ALL_DOCS := \
	content/docs/configuration.md \
	content/docs/installation.md \
	content/docs/operations.md \
	content/docs/workflows/_index.md \
	content/docs/workflows/dau.md

ALL_IMAGES := \
	static/favicons/favicon.ico \
	static/favicons/favicon-16x16.png \
	static/favicons/favicon-32x32.png \
	static/favicons/apple-touch-icon-180x180.png \
	static/favicons/android-36x36.png \
	static/favicons/android-48x48.png \
	static/favicons/android-72x72.png \
	static/favicons/android-96x96.png \
	static/favicons/android-144x144.png \
	static/favicons/android-192x192.png

ALL_FILES := $(ALL_DOCS) $(ALL_IMAGES)

.PHONY: build
build: $(ALL_FILES)
	echo $(ALL_DOCS) | xargs -n1 scripts/postprocess.sh

content/docs/%/_index.md: ../docs/%.md
	$(call convert_markdown)

content/docs/%.md: ../docs/%.md
	$(call convert_markdown)

content/docs/workflows/dau.md: ../workflow/dau/README.md
	$(call convert_markdown)

.PHONY: images
images: $(ALL_IMAGES)

static/favicons/favicon-16x16.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 16x16 $@

static/favicons/favicon.ico static/favicons/favicon-32x32.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 32x32 $@

static/favicons/apple-touch-icon-180x180.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 180x180 $@

static/favicons/android-36x36.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 36x36 $@

static/favicons/android-48x48.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 48x48 $@

static/favicons/android-72x72.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 72x72 $@

static/favicons/android-96x96.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 96x96 $@

static/favicons/android-144x144.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 144x144 $@

static/favicons/android-192x192.png: static/images/logo.svg
	mkdir -p $(@D) && convert $< -transparent white -resize 192x192 $@

.PHONY: clean
clean:
	$(RM) $(ALL_FILES)

define convert_markdown
	./scripts/convert_markdown.sh "$<" "$@"
endef

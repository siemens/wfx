package middleware

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"
)

type GlobalMW struct {
	handler       http.Handler
	intermediates []IntermediateMW
}

type IntermediateMW interface {
	Wrap(next http.Handler) http.Handler
	Shutdown()
}

// implements the http.Handler interface
func (global GlobalMW) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	global.handler.ServeHTTP(rw, r)
}

func (global GlobalMW) Shutdown() {
	for _, mw := range global.intermediates {
		mw.Shutdown()
	}
}

func NewGlobalMiddleware(base http.Handler, intermediates []IntermediateMW) *GlobalMW {
	handler := base
	for _, mw := range intermediates {
		handler = mw.Wrap(handler)
	}
	return &GlobalMW{handler: handler, intermediates: intermediates}
}

type promotedMW struct {
	wrap func(http.Handler) http.Handler
}

func (promotedMW) Shutdown() {}

func (mw promotedMW) Wrap(next http.Handler) http.Handler {
	return mw.wrap(next)
}

func PromoteWrapper(wrap func(http.Handler) http.Handler) IntermediateMW {
	return promotedMW{wrap: wrap}
}

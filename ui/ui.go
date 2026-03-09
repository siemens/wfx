//go:build ui

//go:generate zstd -f -19 dist/app.js
//go:generate openssl dgst -r -sha256 -out dist/app.js.sha256 dist/app.js
//go:generate sed -i -e "s/ .*//" dist/app.js.sha256

//go:generate zstd -f -19 dist/app.css
//go:generate openssl dgst -r -sha256 -out dist/app.css.sha256 dist/app.css
//go:generate sed -i -e "s/ .*//" dist/app.css.sha256

//go:generate zstd -f -19 -o dist/logo.svg.zst ../hugo/static/images/logo.svg

package ui

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog/log"
)

//go:embed index.html.tmpl
var indexTmpl string

var indexHTML []byte

//go:embed dist/app.js.zst
var appJs []byte

//go:embed dist/app.js.sha256
var appJsHash string

//go:embed dist/app.css.zst
var appCss []byte

//go:embed dist/app.css.sha256
var appCssHash string

//go:embed dist/logo.svg.zst
var LogoSVG []byte

type TemplateData struct {
	AppCSS   string
	AppMJS   string
	WfxURL   string
	BasePath string
}

const Enabled = true

const (
	oneDay  = time.Hour * 24
	oneYear = time.Hour * 24 * 365 // approximately
)

func Mux(wfxBasePath string, uiBasePath string) *http.ServeMux {
	appJsHash = strings.TrimSpace(appJsHash)
	appCssHash = strings.TrimSpace(appCssHash)

	tmpl := template.Must(template.New("index").Parse(indexTmpl))

	data := TemplateData{
		AppCSS:   fmt.Sprintf("%s/%s.css", uiBasePath, appCssHash),
		AppMJS:   fmt.Sprintf("%s/%s.mjs", uiBasePath, appJsHash),
		WfxURL:   wfxBasePath,
		BasePath: uiBasePath,
	}

	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, data)
	indexHTML = buf.Bytes()

	log.Info().Msg("Creating UI router")
	mux := http.NewServeMux()
	mux.Handle("/", indexHandler())
	mux.Handle("/logo.svg", zstHandler(LogoSVG, "image/svg+xml", oneDay))
	mux.Handle(fmt.Sprintf("/%s.css", appCssHash), zstHandler(appCss, "text/css", oneYear))
	mux.Handle(fmt.Sprintf("/%s.mjs", appJsHash), zstHandler(appJs, "application/javascript", oneYear))
	return mux
}

func FaviconHandler() http.Handler {
	return zstHandler(LogoSVG, "image/svg+xml", oneDay)
}

func indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// disable caching
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, _ = w.Write(indexHTML)
	})
}

func zstHandler(compressed []byte, contentType string, cacheDuration time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expires := time.Now().Add(cacheDuration).UTC().Format(http.TimeFormat)
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int64(cacheDuration.Seconds())))
		w.Header().Set("Expires", expires)
		w.Header().Set("Content-Type", contentType)

		decoder, _ := zstd.NewReader(nil)
		defer decoder.Close()

		decompressed, err := decoder.DecodeAll(compressed, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		_, _ = w.Write(decompressed)
	})
}

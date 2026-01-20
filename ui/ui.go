//go:build ui

//go:generate zstd -f -19 priv/static/app.min.mjs
//go:generate openssl dgst -r -sha256 -out priv/static/app.min.mjs.sha256 priv/static/app.min.mjs
//go:generate sed -i -e "s/ .*//" priv/static/app.min.mjs.sha256

//go:generate zstd -f -19 priv/static/app.min.css
//go:generate openssl dgst -r -sha256 -out priv/static/app.min.css.sha256 priv/static/app.min.css
//go:generate sed -i -e "s/ .*//" priv/static/app.min.css.sha256

//go:generate zstd -f -19 -o priv/static/logo.svg.zst ../hugo/static/images/logo.svg

package ui

/*
 * SPDX-FileCopyrightText: 2025 Siemens AG
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

//go:embed priv/static/app.min.mjs.zst
var appJs []byte

//go:embed priv/static/app.min.mjs.sha256
var appJsHash string

//go:embed priv/static/app.min.css.zst
var appCss []byte

//go:embed priv/static/app.min.css.sha256
var appCssHash string

//go:embed priv/static/logo.svg.zst
var LogoSVG []byte

type TemplateData struct {
	AppCSS   string
	AppMJS   string
	WfxURL   string
	BasePath string
}

const Enabled = true

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
	mux.Handle("/logo.svg", zstHandler(LogoSVG, "image/svg+xml"))
	mux.Handle(fmt.Sprintf("/%s.css", appCssHash), zstHandler(appCss, "text/css"))
	mux.Handle(fmt.Sprintf("/%s.mjs", appJsHash), zstHandler(appJs, "application/javascript"))
	return mux
}

func FaviconHandler() http.Handler {
	return zstHandler(LogoSVG, "image/svg+xml")
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

func zstHandler(compressed []byte, contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow client to cache content for up to 1 year
		expires := time.Now().AddDate(1, 0, 0).UTC().Format(http.TimeFormat)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
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

// Copyright © 2023 Aspen James <hello@aspenjames.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the “Software”), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"compress/gzip"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

var (
	//go:embed content/static/* content/templates/*
	content   embed.FS
	static    fs.FS
	mimeTypes map[string]string = map[string]string{
		".css":  "text/css",
		".ico":  "image/x-icon",
		".js":   "text/javascript",
		".pdf":  "application/pdf",
		".wasm": "application/wasm",
	}

	err404Tmpl *template.Template
	templates  *template.Template
	pages      map[string]string = map[string]string{
		"/":          "content/templates/index.html",
		"/about":     "content/templates/about.html",
		"/particles": "content/templates/particles.html",
		"/resume":    "content/templates/resume.html",
	}

	//go:embed content/routes.json
	routes []byte
	links  []*navLink

	// Default max-age 1 week; 1 day for 404.
	cacheMaxAge       int    = 60 * 60 * 24 * 7
	cacheMaxAge404    int    = 60 * 60 * 24
	darkModeCookieKey string = "aj-dot-dev##dark-mode"
)

type navLink struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Active bool
}

type tmplData struct {
	DarkMode          bool
	DarkModeCookieKey string
	NavLinks          []*navLink
}

// Returns the tmplData for the current request.
func getTmplData(r *fsthttp.Request) tmplData {
	for _, l := range links {
		l.Active = l.Path == r.URL.Path
	}

	dmCookie, err := r.Cookie(darkModeCookieKey)
	if err != nil {
		dmCookie = &fsthttp.Cookie{
			Name:     darkModeCookieKey,
			Domain:   r.URL.Hostname(),
			SameSite: fsthttp.SameSiteStrictMode,
			Expires:  time.Now().Add(time.Hour * 24 * 365).UTC(),
			Value:    "light",
		}
		r.AddCookie(dmCookie)
	}

	return tmplData{
		DarkModeCookieKey: darkModeCookieKey,
		DarkMode:          dmCookie.Value == "dark",
		NavLinks:          links,
	}
}

func init() {
	var err error
	if err = json.Unmarshal(routes, &links); err != nil {
		log.Fatal(err)
	}

	if static, err = fs.Sub(content, "content/static"); err != nil {
		log.Fatal(err)
	}
	if templates, err = template.ParseFS(
		content,
		"content/templates/partials/*.html",
		"content/templates/layouts/main.html",
	); err != nil {
		log.Fatal(err)
	}
	if err404Tmpl, err = template.ParseFS(
		content,
		"content/templates/layouts/error.html",
		"content/templates/errors/404.html",
	); err != nil {
		log.Fatal(err)
	}
}

func canCompress(r *fsthttp.Request) bool {
	for _, ac := range r.Header["Accept-Encoding"] {
		if ac == "gzip" {
			return true
		}
	}
	return false
}

func main() {
	fsthttp.ServeFunc(func(ctx context.Context, w fsthttp.ResponseWriter, r *fsthttp.Request) {
		// Filter requests that have unexpected methods.
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE" {
			w.WriteHeader(fsthttp.StatusMethodNotAllowed)
			fmt.Fprintf(w, "This method is not allowed\n")
			return
		}
		// Route requests.
		switch {
		case r.URL.Path == "/favicon.ico":
			// Serve favicon from static fs.
			if favicon, err := fs.ReadFile(static, "favicon.ico"); err != nil {
				w.WriteHeader(fsthttp.StatusNotFound)
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAge404))
			} else {
				w.Header().Set("Content-Type", "image/x-icon")
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAge))
				if r.Method == "GET" {
					if canCompress(r) {
						w.Header().Set("Content-Encoding", "gzip")
						gw := gzip.NewWriter(w)
						defer gw.Close()

						fmt.Fprint(gw, string(favicon))
					} else {
						fmt.Fprint(w, string(favicon))
					}
				}
			}
		case strings.HasPrefix(r.URL.Path, "/static"):
			// Serve static content, if found.
			if fdata, err := fs.ReadFile(static, strings.TrimPrefix(r.URL.Path, "/static/")); err != nil {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAge404))
				w.WriteHeader(fsthttp.StatusNotFound)
			} else {
				content_type, ok := mimeTypes[filepath.Ext(r.URL.Path)]
				if !ok {
					// Fallback to octect-stream for safety
					// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
					content_type = "application/octect-stream"
				}
				w.Header().Set("Content-Type", content_type)
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAge))
				if r.Method == "GET" {
					if canCompress(r) {
						w.Header().Set("Content-Encoding", "gzip")
						gw := gzip.NewWriter(w)
						defer gw.Close()

						fmt.Fprint(gw, string(fdata))
					} else {
						fmt.Fprint(w, string(fdata))
					}
				}
			}
		default:
			// Render page template.
			tmplData := getTmplData(r)
			w.Header().Set("Content-Type", "text/html")
			pageTmpl, ok := pages[r.URL.Path]
			if !ok {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAge404))
				w.WriteHeader(fsthttp.StatusNotFound)
				if r.Method == "GET" {
					if canCompress(r) {
						w.Header().Set("Content-Encoding", "gzip")
						gw := gzip.NewWriter(w)
						defer gw.Close()

						if err := err404Tmpl.ExecuteTemplate(gw, "404.html", tmplData); err != nil {
							w.WriteHeader(fsthttp.StatusInternalServerError)
							log.Fatal(err)
						}
					} else {
						if err := err404Tmpl.ExecuteTemplate(w, "404.html", tmplData); err != nil {
							w.WriteHeader(fsthttp.StatusInternalServerError)
							log.Fatal(err)
						}
					}
				}
			} else {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheMaxAge))
				if r.Method == "GET" {
					template.Must(templates.ParseFS(content, pageTmpl))
					if canCompress(r) {
						w.Header().Set("Content-Encoding", "gzip")
						gw := gzip.NewWriter(w)
						defer gw.Close()

						if err := templates.ExecuteTemplate(gw, filepath.Base(pageTmpl), tmplData); err != nil {
							w.WriteHeader(fsthttp.StatusInternalServerError)
							log.Fatal(err)
						}
					} else {
						if err := templates.ExecuteTemplate(w, filepath.Base(pageTmpl), tmplData); err != nil {
							w.WriteHeader(fsthttp.StatusInternalServerError)
							log.Fatal(err)
						}
					}
				}
			}
		}
	})
}

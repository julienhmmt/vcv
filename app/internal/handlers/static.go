package handlers

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// RegisterStaticRoutes serves the embedded Vite build (dist) as an SPA-ish
// multi-page app. `/` returns dist/index.html, `/admin` returns dist/admin.html,
// `/assets/*` is served as-is, and other static files at the root (favicons,
// manifest, svg icons) are resolved directly from dist.
func RegisterStaticRoutes(router chi.Router, distFS fs.FS) {
	fileServer := http.FileServer(http.FS(distFS))

	serveFile := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data, err := fs.ReadFile(distFS, name)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(data)
		}
	}

	router.Get("/", serveFile("index.html"))
	router.Get("/admin", serveFile("admin.html"))

	router.Handle("/assets/*", immutableCache(fileServer))

	// Root-level static files emitted by Vite from public/
	rootFiles := []string{
		"favicon.ico",
		"favicon.svg",
		"favicon-16x16.png",
		"favicon-32x32.png",
		"apple-touch-icon.png",
		"android-chrome-192x192.png",
		"android-chrome-512x512.png",
		"site.webmanifest",
		"icons.svg",
		"github.svg",
		"docker.svg",
	}
	for _, name := range rootFiles {
		path := "/" + name
		fileName := strings.TrimPrefix(path, "/")
		router.Get(path, serveFile(fileName))
	}
}

// immutableCache sets a long-lived Cache-Control header on hashed assets.
// Vite emits filenames like `index-DJxyvc_i.js`, so content changes always
// change the URL and stale caches cannot serve a wrong file.
func immutableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

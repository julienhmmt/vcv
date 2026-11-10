package handlers

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// contentTypeFor returns the Content-Type for a served file. Go's built-in mime
// table is used first; the explicit entries cover types missing from some
// systems' mime tables (CI images) and types Go does not map.
// shortcut: explicit map covers only the extensions this app serves, add new ones here if needed.
func contentTypeFor(name string) string {
	switch ext := filepath.Ext(name); ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".webmanifest":
		return "application/manifest+json"
	case ".ico":
		return "image/x-icon"
	default:
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
		return "application/octet-stream"
	}
}

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
			w.Header().Set("Content-Type", contentTypeFor(name))
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

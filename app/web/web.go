package web

import "embed"

// EmbeddedFS contains the compiled-in Svelte frontend (dist).
// Populated by `task web-build` (Vite). Run it before `go build`.
//
//go:embed all:dist
var EmbeddedFS embed.FS

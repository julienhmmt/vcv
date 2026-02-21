package web

import "embed"

// EmbeddedFS contains the compiled-in static assets for the UI.
//
//go:embed *
var EmbeddedFS embed.FS

package main

import "embed"

// embeddedWeb contains the compiled-in static assets for the UI.
//
//go:embed web/*
var embeddedWeb embed.FS

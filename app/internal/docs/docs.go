// Package docs embeds VCV's user-facing documentation and renders it to HTML.
package docs

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:embed ADMIN.md
var adminMarkdown string

var (
	adminHTMLOnce sync.Once
	adminHTML     string
)

// AdminHTML renders the embedded admin guide to HTML. The markdown source is
// compiled into the binary and authored in-repo, so the output is trusted and
// rendered once, lazily.
func AdminHTML() string {
	adminHTMLOnce.Do(func() {
		md := goldmark.New(goldmark.WithExtensions(extension.GFM))
		var buf bytes.Buffer
		if err := md.Convert([]byte(adminMarkdown), &buf); err != nil {
			adminHTML = "<p>Documentation unavailable.</p>"
			return
		}
		adminHTML = buf.String()
	})
	return adminHTML
}

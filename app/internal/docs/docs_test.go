package docs

import (
	"strings"
	"testing"
)

func TestAdminHTML(t *testing.T) {
	html := AdminHTML()
	if html == "" {
		t.Fatal("AdminHTML returned empty string")
	}
	for _, want := range []string{"<h1", "<h2", "<code", "Admin Guide"} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}

func TestAdminHTMLCached(t *testing.T) {
	first := AdminHTML()
	second := AdminHTML()
	if first != second {
		t.Fatal("AdminHTML should return stable cached output")
	}
}

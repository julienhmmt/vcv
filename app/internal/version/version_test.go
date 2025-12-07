package version

import "testing"

func TestInfoContainsVersion(t *testing.T) {
	info := Info()
	if info["version"] == "" {
		t.Fatalf("expected version to be present")
	}
}

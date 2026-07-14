package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	var buf bytes.Buffer
	PrintVersion(&buf, BuildInfo{Version: "v0.1.0", Commit: "abc123", Date: "2026-07-15T12:00:00Z"})
	got := buf.String()
	for _, want := range []string{"quickserve v0.1.0", "commit: abc123", "built:  2026-07-15T12:00:00Z"} {
		if !strings.Contains(got, want) {
			t.Fatalf("version output missing %q: %s", want, got)
		}
	}
}

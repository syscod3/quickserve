package tunnel

import (
	"strings"
	"testing"
)

func TestScanForURLFindsTryCloudflareURL(t *testing.T) {
	input := strings.NewReader("info\nYour quick Tunnel has been created! Visit it at https://blue-river-123.trycloudflare.com\n")
	got, err := scanForURL(input, "")
	if err != nil {
		t.Fatalf("scanForURL() error = %v", err)
	}
	if got != "https://blue-river-123.trycloudflare.com" {
		t.Fatalf("scanForURL() = %q", got)
	}
}

func TestScanForURLReportsMissingURLWithOutput(t *testing.T) {
	_, err := scanForURL(strings.NewReader("failed to connect\n"), "")
	if err == nil {
		t.Fatal("scanForURL() succeeded unexpectedly")
	}
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("error does not include recent output: %v", err)
	}
}

func TestScanForURLReturnsCustomHostnameAfterReadyLine(t *testing.T) {
	input := strings.NewReader("INF Registered tunnel connection connIndex=0\n")
	got, err := scanForURL(input, "share.example.com")
	if err != nil {
		t.Fatalf("scanForURL() error = %v", err)
	}
	if got != "https://share.example.com" {
		t.Fatalf("scanForURL() = %q", got)
	}
}

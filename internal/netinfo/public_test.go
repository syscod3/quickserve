package netinfo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLookupPublicIPv4ReturnsValidAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("8.8.8.8\n"))
	}))
	defer server.Close()

	got, err := LookupPublicIPv4(context.Background(), server.Client(), server.URL, 1024)
	if err != nil {
		t.Fatalf("LookupPublicIPv4() error = %v", err)
	}
	if got != "8.8.8.8" {
		t.Fatalf("LookupPublicIPv4() = %q", got)
	}
}

func TestLookupPublicIPv4RejectsMalformedPrivateNon200AndOversized(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		limit  int64
	}{
		{"malformed", http.StatusOK, "x", 1024},
		{"private", http.StatusOK, "192.168.1.1", 1024},
		{"non-200", http.StatusTeapot, "8.8.8.8", 1024},
		{"oversized", http.StatusOK, "8.8.8.8\nextra", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			_, err := LookupPublicIPv4(context.Background(), server.Client(), server.URL, tt.limit)
			if err == nil {
				t.Fatal("LookupPublicIPv4() succeeded unexpectedly")
			}
		})
	}
}

func TestLookupPublicIPv4HonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte("8.8.8.8"))
	}))
	defer server.Close()

	client := server.Client()
	client.Timeout = 50 * time.Millisecond
	_, err := LookupPublicIPv4(context.Background(), client, server.URL, 1024)
	if err == nil {
		t.Fatal("LookupPublicIPv4() succeeded unexpectedly")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Logf("timeout surfaced as: %v", err)
	}
}

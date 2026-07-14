package netinfo

import (
	"context"
	"net/http"
	"time"
)

type Provider struct {
	Endpoint string
	Client   *http.Client
	Limit    int64
}

func DefaultProvider() Provider {
	return Provider{
		Endpoint: "https://api.ipify.org",
		Client:   &http.Client{Timeout: 3 * time.Second},
		Limit:    128,
	}
}

func (p Provider) LANIPv4(ctx context.Context) (string, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return LANIPv4(lookupCtx)
}

func (p Provider) PublicIPv4(ctx context.Context) (string, error) {
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = "https://api.ipify.org"
	}
	limit := p.Limit
	if limit == 0 {
		limit = 128
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return LookupPublicIPv4(lookupCtx, p.Client, endpoint, limit)
}

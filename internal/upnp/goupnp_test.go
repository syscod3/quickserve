package upnp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDiscoverKeepsSearchingAfterOneServiceTimesOut(t *testing.T) {
	client := &fakeClient{external: "8.8.8.8"}
	searches := []igdSearch{
		func(ctx context.Context) ([]Client, []error, error) {
			<-ctx.Done()
			return nil, nil, ctx.Err()
		},
		func(searchCtx context.Context) ([]Client, []error, error) {
			if err := searchCtx.Err(); err != nil {
				return nil, nil, err
			}
			return []Client{client}, nil, nil
		},
	}

	discovery := IGDDiscovery{Timeout: 10 * time.Millisecond}
	clients, err := discovery.discover(context.Background(), searches)
	if err != nil {
		t.Fatalf("discover() error = %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("discover() returned %d clients, want 1", len(clients))
	}
}

func TestDiscoverReportsErrorsWhenNoServicesMatch(t *testing.T) {
	discovery := IGDDiscovery{Timeout: time.Second}
	_, err := discovery.discover(context.Background(), []igdSearch{
		func(context.Context) ([]Client, []error, error) {
			return nil, nil, errors.New("no route")
		},
	})
	if err == nil {
		t.Fatal("discover() succeeded unexpectedly")
	}
}

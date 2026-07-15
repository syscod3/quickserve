package upnp

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerCreatesRenewsAndCleansMapping(t *testing.T) {
	client := &fakeClient{external: "8.8.8.8"}
	m := NewManager(fakeDiscovery{clients: []Client{client}}, WithRenewTick(10*time.Millisecond))

	mapping, err := m.Map(context.Background(), Request{
		LocalIP:      net.ParseIP("192.168.1.10"),
		LocalPort:    8000,
		ExternalPort: 18080,
		Lease:        30 * time.Millisecond,
		Description:  "quickserve",
	})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if mapping.ExternalIP != "8.8.8.8" {
		t.Fatalf("ExternalIP = %q", mapping.ExternalIP)
	}
	time.Sleep(35 * time.Millisecond)
	if client.adds.Load() < 2 {
		t.Fatalf("renewal did not occur, adds = %d", client.adds.Load())
	}
	if err := mapping.Cleanup(context.Background()); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
	if err := mapping.Cleanup(context.Background()); err != nil {
		t.Fatalf("second Cleanup() error = %v", err)
	}
	if client.deletes.Load() != 1 {
		t.Fatalf("deletes = %d, want 1", client.deletes.Load())
	}
}

func TestManagerReportsDiscoveryFailure(t *testing.T) {
	m := NewManager(fakeDiscovery{err: errors.New("offline")})
	_, err := m.Map(context.Background(), Request{LocalIP: net.ParseIP("192.168.1.10"), LocalPort: 8000})
	if err == nil {
		t.Fatal("Map() succeeded unexpectedly")
	}
}

func TestManagerExplainsMissingIGD(t *testing.T) {
	m := NewManager(fakeDiscovery{})
	_, err := m.Map(context.Background(), Request{LocalIP: net.ParseIP("192.168.1.10"), LocalPort: 8000})
	if err == nil {
		t.Fatal("Map() succeeded unexpectedly")
	}
	for _, want := range []string{
		"no UPnP Internet Gateway Device",
		"WANIPConnection",
		"router that performs NAT",
		"double NAT",
		"manual port forward",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}

func TestManagerTriesAllClients(t *testing.T) {
	reject := &fakeClient{addErr: errors.New("conflict")}
	accept := &fakeClient{external: "8.8.4.4"}
	m := NewManager(fakeDiscovery{clients: []Client{reject, accept}})

	mapping, err := m.Map(context.Background(), Request{LocalIP: net.ParseIP("192.168.1.10"), LocalPort: 8000})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	defer mapping.Cleanup(context.Background())
	if reject.adds.Load() != 1 || accept.adds.Load() != 1 {
		t.Fatalf("adds reject=%d accept=%d", reject.adds.Load(), accept.adds.Load())
	}
}

func TestManagerDoesNotCleanupBeforeSuccessfulCreate(t *testing.T) {
	client := &fakeClient{addErr: errors.New("reject")}
	m := NewManager(fakeDiscovery{clients: []Client{client}})
	_, err := m.Map(context.Background(), Request{LocalIP: net.ParseIP("192.168.1.10"), LocalPort: 8000})
	if err == nil {
		t.Fatal("Map() succeeded unexpectedly")
	}
	if client.deletes.Load() != 0 {
		t.Fatalf("deletes = %d, want 0", client.deletes.Load())
	}
}

type fakeDiscovery struct {
	clients []Client
	err     error
}

func (f fakeDiscovery) Discover(context.Context) ([]Client, error) {
	return f.clients, f.err
}

type fakeClient struct {
	adds     atomic.Int64
	deletes  atomic.Int64
	addErr   error
	delErr   error
	external string
}

func (f *fakeClient) AddPortMapping(context.Context, MappingSpec) error {
	f.adds.Add(1)
	return f.addErr
}

func (f *fakeClient) DeletePortMapping(context.Context, uint16, string) error {
	f.deletes.Add(1)
	return f.delErr
}

func (f *fakeClient) ExternalIP(context.Context) (string, error) {
	return f.external, nil
}

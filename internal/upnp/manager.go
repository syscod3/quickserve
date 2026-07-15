package upnp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const noIGDMessage = "no UPnP Internet Gateway Device (IGD) was found. UPnP may be present on the LAN, but quickserve needs a NAT router exposing WANIPConnection or WANPPPConnection. Enable UPnP/IGD on the router that performs NAT, or create a manual port forward. If this router's WAN is behind another ISP router, double NAT means the ISP router also needs forwarding or bridge/modem mode"

type Client interface {
	AddPortMapping(context.Context, MappingSpec) error
	DeletePortMapping(context.Context, uint16, string) error
	ExternalIP(context.Context) (string, error)
}

type Discovery interface {
	Discover(context.Context) ([]Client, error)
}

type MappingSpec struct {
	LocalIP      net.IP
	LocalPort    uint16
	ExternalPort uint16
	Protocol     string
	Description  string
	Lease        time.Duration
}

type Request struct {
	LocalIP      net.IP
	LocalPort    int
	ExternalPort int
	Lease        time.Duration
	Description  string
}

type Manager struct {
	discovery  Discovery
	renewTick  time.Duration
	protocol   string
	defaultTag string
}

type Option func(*Manager)

func WithRenewTick(d time.Duration) Option {
	return func(m *Manager) {
		m.renewTick = d
	}
}

func NewManager(discovery Discovery, opts ...Option) *Manager {
	m := &Manager{
		discovery:  discovery,
		renewTick:  time.Minute,
		protocol:   "TCP",
		defaultTag: "quickserve",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Mapping struct {
	ExternalIP   string
	ExternalPort uint16

	client Client
	spec   MappingSpec
	cancel context.CancelFunc
	once   sync.Once
	err    error
}

func (m *Manager) Map(ctx context.Context, req Request) (*Mapping, error) {
	if m.discovery == nil {
		return nil, errors.New("UPnP discovery is not configured")
	}
	clients, err := m.discovery.Discover(ctx)
	if err != nil {
		return nil, fmt.Errorf("discover UPnP IGD: %w", err)
	}
	if len(clients) == 0 {
		return nil, errors.New(noIGDMessage)
	}
	spec, err := requestSpec(req, m.protocol, m.defaultTag)
	if err != nil {
		return nil, err
	}
	var failures []error
	for _, client := range clients {
		if err := client.AddPortMapping(ctx, spec); err != nil {
			failures = append(failures, err)
			continue
		}
		external, _ := client.ExternalIP(ctx)
		renewCtx, cancel := context.WithCancel(context.Background())
		mapping := &Mapping{
			ExternalIP:   external,
			ExternalPort: spec.ExternalPort,
			client:       client,
			spec:         spec,
			cancel:       cancel,
		}
		if spec.Lease > 0 {
			go m.renew(renewCtx, client, spec)
		}
		return mapping, nil
	}
	return nil, fmt.Errorf("all UPnP mapping attempts failed: %w", errors.Join(failures...))
}

func (m *Manager) renew(ctx context.Context, client Client, spec MappingSpec) {
	interval := spec.Lease / 2
	if interval <= 0 || interval > m.renewTick {
		interval = m.renewTick
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = client.AddPortMapping(ctx, spec)
		}
	}
}

func (m *Mapping) Cleanup(ctx context.Context) error {
	m.once.Do(func() {
		m.cancel()
		m.err = m.client.DeletePortMapping(ctx, m.spec.ExternalPort, m.spec.Protocol)
	})
	return m.err
}

func requestSpec(req Request, protocol, defaultDescription string) (MappingSpec, error) {
	if req.LocalIP == nil || req.LocalIP.To4() == nil {
		return MappingSpec{}, errors.New("local IPv4 address is required for UPnP mapping")
	}
	if req.LocalPort <= 0 || req.LocalPort > 65535 {
		return MappingSpec{}, fmt.Errorf("local port %d is invalid", req.LocalPort)
	}
	external := req.ExternalPort
	if external == 0 {
		external = req.LocalPort
	}
	if external <= 0 || external > 65535 {
		return MappingSpec{}, fmt.Errorf("external port %d is invalid", external)
	}
	if req.Lease < 0 {
		return MappingSpec{}, errors.New("UPnP lease duration must not be negative")
	}
	description := req.Description
	if description == "" {
		description = defaultDescription
	}
	return MappingSpec{
		LocalIP:      req.LocalIP,
		LocalPort:    uint16(req.LocalPort),
		ExternalPort: uint16(external),
		Protocol:     protocol,
		Description:  description,
		Lease:        req.Lease,
	}, nil
}

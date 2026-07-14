package upnp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	igd1 "github.com/huin/goupnp/dcps/internetgateway1"
	igd2 "github.com/huin/goupnp/dcps/internetgateway2"
)

type IGDDiscovery struct {
	Timeout time.Duration
}

func (d IGDDiscovery) Discover(ctx context.Context) ([]Client, error) {
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var clients []Client
	var errs []error
	add := func(found []Client, discovered []error, err error) {
		clients = append(clients, found...)
		errs = append(errs, discovered...)
		if err != nil {
			errs = append(errs, err)
		}
	}

	c2, e2, err2 := igd2.NewWANIPConnection2ClientsCtx(ctx)
	var w2 []Client
	for _, c := range c2 {
		w2 = append(w2, wanIP2{client: c, local: c.LocalAddr()})
	}
	add(w2, e2, err2)

	c21, e21, err21 := igd2.NewWANIPConnection1ClientsCtx(ctx)
	var w21 []Client
	for _, c := range c21 {
		w21 = append(w21, wanIP21{client: c, local: c.LocalAddr()})
	}
	add(w21, e21, err21)

	c1, e1, err1 := igd1.NewWANIPConnection1ClientsCtx(ctx)
	var w1 []Client
	for _, c := range c1 {
		w1 = append(w1, wanIP1{client: c, local: c.LocalAddr()})
	}
	add(w1, e1, err1)

	ppp2, ep2, errp2 := igd2.NewWANPPPConnection1ClientsCtx(ctx)
	var wp2 []Client
	for _, c := range ppp2 {
		wp2 = append(wp2, wanPPP2{client: c, local: c.LocalAddr()})
	}
	add(wp2, ep2, errp2)

	ppp1, ep1, errp1 := igd1.NewWANPPPConnection1ClientsCtx(ctx)
	var wp1 []Client
	for _, c := range ppp1 {
		wp1 = append(wp1, wanPPP1{client: c, local: c.LocalAddr()})
	}
	add(wp1, ep1, errp1)

	if len(clients) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return clients, nil
}

type wanIP1 struct {
	client *igd1.WANIPConnection1
	local  net.IP
}

func (w wanIP1) AddPortMapping(ctx context.Context, spec MappingSpec) error {
	return w.client.AddPortMappingCtx(ctx, "", spec.ExternalPort, spec.Protocol, spec.LocalPort, localClient(w.local, spec.LocalIP), true, spec.Description, leaseSeconds(spec.Lease))
}
func (w wanIP1) DeletePortMapping(ctx context.Context, port uint16, protocol string) error {
	return w.client.DeletePortMappingCtx(ctx, "", port, protocol)
}
func (w wanIP1) ExternalIP(ctx context.Context) (string, error) {
	return w.client.GetExternalIPAddressCtx(ctx)
}

type wanIP21 struct {
	client *igd2.WANIPConnection1
	local  net.IP
}

func (w wanIP21) AddPortMapping(ctx context.Context, spec MappingSpec) error {
	return w.client.AddPortMappingCtx(ctx, "", spec.ExternalPort, spec.Protocol, spec.LocalPort, localClient(w.local, spec.LocalIP), true, spec.Description, leaseSeconds(spec.Lease))
}
func (w wanIP21) DeletePortMapping(ctx context.Context, port uint16, protocol string) error {
	return w.client.DeletePortMappingCtx(ctx, "", port, protocol)
}
func (w wanIP21) ExternalIP(ctx context.Context) (string, error) {
	return w.client.GetExternalIPAddressCtx(ctx)
}

type wanIP2 struct {
	client *igd2.WANIPConnection2
	local  net.IP
}

func (w wanIP2) AddPortMapping(ctx context.Context, spec MappingSpec) error {
	return w.client.AddPortMappingCtx(ctx, "", spec.ExternalPort, spec.Protocol, spec.LocalPort, localClient(w.local, spec.LocalIP), true, spec.Description, leaseSeconds(spec.Lease))
}
func (w wanIP2) DeletePortMapping(ctx context.Context, port uint16, protocol string) error {
	return w.client.DeletePortMappingCtx(ctx, "", port, protocol)
}
func (w wanIP2) ExternalIP(ctx context.Context) (string, error) {
	return w.client.GetExternalIPAddressCtx(ctx)
}

type wanPPP1 struct {
	client *igd1.WANPPPConnection1
	local  net.IP
}

func (w wanPPP1) AddPortMapping(ctx context.Context, spec MappingSpec) error {
	return w.client.AddPortMappingCtx(ctx, "", spec.ExternalPort, spec.Protocol, spec.LocalPort, localClient(w.local, spec.LocalIP), true, spec.Description, leaseSeconds(spec.Lease))
}
func (w wanPPP1) DeletePortMapping(ctx context.Context, port uint16, protocol string) error {
	return w.client.DeletePortMappingCtx(ctx, "", port, protocol)
}
func (w wanPPP1) ExternalIP(ctx context.Context) (string, error) {
	return w.client.GetExternalIPAddressCtx(ctx)
}

type wanPPP2 struct {
	client *igd2.WANPPPConnection1
	local  net.IP
}

func (w wanPPP2) AddPortMapping(ctx context.Context, spec MappingSpec) error {
	return w.client.AddPortMappingCtx(ctx, "", spec.ExternalPort, spec.Protocol, spec.LocalPort, localClient(w.local, spec.LocalIP), true, spec.Description, leaseSeconds(spec.Lease))
}
func (w wanPPP2) DeletePortMapping(ctx context.Context, port uint16, protocol string) error {
	return w.client.DeletePortMappingCtx(ctx, "", port, protocol)
}
func (w wanPPP2) ExternalIP(ctx context.Context) (string, error) {
	return w.client.GetExternalIPAddressCtx(ctx)
}

func localClient(preferred, fallback net.IP) string {
	if preferred != nil && preferred.To4() != nil {
		return preferred.String()
	}
	return fallback.String()
}

func leaseSeconds(d time.Duration) uint32 {
	if d <= 0 {
		return 0
	}
	seconds := uint64(d.Round(time.Second) / time.Second)
	if seconds > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(seconds)
}

func NewDefaultManager() *Manager {
	return NewManager(IGDDiscovery{Timeout: 5 * time.Second}, WithRenewTick(5*time.Minute))
}

func (r Request) String() string {
	return fmt.Sprintf("%s:%d -> %d/%s", r.LocalIP, r.LocalPort, r.ExternalPort, "TCP")
}

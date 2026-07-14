package netinfo

import (
	"context"
	"net"
)

func LANIPv4(ctx context.Context) (string, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp4", "1.1.1.1:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP == nil {
		return "", nil
	}
	return addr.IP.String(), nil
}

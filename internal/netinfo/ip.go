package netinfo

import "net"

func IsGlobalIPv4(value string) bool {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}
	ip = ip.To4()
	if ip == nil {
		return false
	}
	if ip[0] == 0 || ip[0] == 10 || ip[0] == 127 || ip[0] >= 224 {
		return false
	}
	if ip[0] == 100 && ip[1] >= 64 && ip[1] <= 127 {
		return false
	}
	if ip[0] == 169 && ip[1] == 254 {
		return false
	}
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		return false
	}
	if ip[0] == 192 && ip[1] == 168 {
		return false
	}
	return true
}

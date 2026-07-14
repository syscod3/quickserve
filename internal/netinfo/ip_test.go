package netinfo

import "testing"

func TestIsGlobalIPv4(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":         true,
		"1.1.1.1":         true,
		"10.0.0.1":        false,
		"172.16.0.1":      false,
		"192.168.1.1":     false,
		"127.0.0.1":       false,
		"169.254.1.1":     false,
		"224.0.0.1":       false,
		"0.0.0.0":         false,
		"100.64.0.1":      false,
		"100.127.255.254": false,
		"not-an-ip":       false,
		"2001:4860::8888": false,
	}

	for input, want := range cases {
		if got := IsGlobalIPv4(input); got != want {
			t.Fatalf("IsGlobalIPv4(%q) = %v, want %v", input, got, want)
		}
	}
}

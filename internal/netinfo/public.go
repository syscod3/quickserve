package netinfo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrNonGlobalAddress = errors.New("address is not a globally routable IPv4 address")

func LookupPublicIPv4(ctx context.Context, client *http.Client, endpoint string, maxBytes int64) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if maxBytes <= 0 {
		return "", fmt.Errorf("max bytes must be positive")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("public IP lookup returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(body)) > maxBytes {
		return "", fmt.Errorf("public IP response exceeded %d bytes", maxBytes)
	}
	value := strings.TrimSpace(string(body))
	if !IsGlobalIPv4(value) {
		return "", ErrNonGlobalAddress
	}
	return value, nil
}

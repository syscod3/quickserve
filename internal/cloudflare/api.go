package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DefaultBaseURL = "https://api.cloudflare.com/client/v4"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Zone struct {
	ID      string
	Name    string
	Account Account
}

type Account struct {
	ID   string
	Name string
}

type Tunnel struct {
	ID     string
	Name   string
	Status string
}

type IngressRule struct {
	Hostname string `json:"hostname,omitempty"`
	Service  string `json:"service"`
}

type DNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl,omitempty"`
}

func (c Client) TunnelToken(ctx context.Context, accountID, tunnelID, apiToken string) (string, error) {
	if accountID == "" {
		return "", fmt.Errorf("account id is required")
	}
	if tunnelID == "" {
		return "", fmt.Errorf("tunnel id is required")
	}
	if apiToken == "" {
		return "", fmt.Errorf("api token is required")
	}

	body, err := c.get(ctx, fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/token", accountID, tunnelID), nil, apiToken)
	if err != nil {
		return "", err
	}
	return parseTunnelToken(body)
}

func (c Client) ZoneByName(ctx context.Context, name, apiToken string) (Zone, error) {
	if name == "" {
		return Zone{}, fmt.Errorf("zone name is required")
	}
	body, err := c.get(ctx, "/zones", map[string]string{"name": name, "per_page": "1"}, apiToken)
	if err != nil {
		return Zone{}, err
	}
	var envelope struct {
		Result []Zone `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return Zone{}, err
	}
	if len(envelope.Result) == 0 {
		return Zone{}, fmt.Errorf("zone %q was not found", name)
	}
	return envelope.Result[0], nil
}

func (c Client) FindZoneForHostname(ctx context.Context, hostname, apiToken string) (Zone, error) {
	parts := strings.Split(strings.Trim(hostname, "."), ".")
	for i := 0; i < len(parts)-1; i++ {
		name := strings.Join(parts[i:], ".")
		zone, err := c.ZoneByName(ctx, name, apiToken)
		if err == nil {
			return zone, nil
		}
		if !strings.Contains(err.Error(), "was not found") {
			return Zone{}, err
		}
	}
	return Zone{}, fmt.Errorf("no Cloudflare zone found for hostname %q", hostname)
}

func (c Client) Tunnels(ctx context.Context, accountID, name, apiToken string) ([]Tunnel, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account id is required")
	}
	query := map[string]string{"per_page": "100"}
	if name != "" {
		query["name"] = name
	}
	body, err := c.get(ctx, fmt.Sprintf("/accounts/%s/tunnels", accountID), query, apiToken)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Result []Tunnel `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return envelope.Result, nil
}

func (c Client) TunnelIngress(ctx context.Context, accountID, tunnelID, apiToken string) ([]IngressRule, error) {
	body, err := c.get(ctx, fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", accountID, tunnelID), nil, apiToken)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Result struct {
			Config struct {
				Ingress []IngressRule `json:"ingress"`
			} `json:"config"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return envelope.Result.Config.Ingress, nil
}

func (c Client) PutTunnelIngress(ctx context.Context, accountID, tunnelID string, ingress []IngressRule, apiToken string) error {
	_, err := c.put(ctx, fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", accountID, tunnelID), map[string]any{
		"config": map[string]any{"ingress": ingress},
	}, apiToken)
	return err
}

func UpsertTunnelIngress(ingress []IngressRule, hostname, service string) []IngressRule {
	for i := range ingress {
		if ingress[i].Hostname == hostname {
			ingress[i].Service = service
			return ensureFallbackIngress(ingress)
		}
	}
	rule := IngressRule{Hostname: hostname, Service: service}
	for i, existing := range ingress {
		if existing.Hostname == "" && strings.HasPrefix(existing.Service, "http_status:") {
			next := append([]IngressRule{}, ingress[:i]...)
			next = append(next, rule)
			next = append(next, ingress[i:]...)
			return ensureFallbackIngress(next)
		}
	}
	ingress = append(ingress, rule)
	return ensureFallbackIngress(ingress)
}

func ensureFallbackIngress(ingress []IngressRule) []IngressRule {
	for _, rule := range ingress {
		if rule.Hostname == "" && strings.HasPrefix(rule.Service, "http_status:") {
			return ingress
		}
	}
	return append(ingress, IngressRule{Service: "http_status:404"})
}

func (c Client) UpsertDNSCNAME(ctx context.Context, zoneID, name, content, apiToken string) (DNSRecord, error) {
	records, err := c.dnsRecords(ctx, zoneID, name, "CNAME", apiToken)
	if err != nil {
		return DNSRecord{}, err
	}
	record := DNSRecord{Type: "CNAME", Name: name, Content: content, Proxied: true, TTL: 1}
	if len(records) > 0 {
		record.ID = records[0].ID
		body, err := c.put(ctx, fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, record.ID), record, apiToken)
		if err != nil {
			return DNSRecord{}, err
		}
		return parseDNSRecord(body)
	}
	body, err := c.post(ctx, fmt.Sprintf("/zones/%s/dns_records", zoneID), record, apiToken)
	if err != nil {
		return DNSRecord{}, err
	}
	return parseDNSRecord(body)
}

func parseTunnelToken(body []byte) (string, error) {
	var token string
	if err := json.Unmarshal(body, &token); err == nil && token != "" {
		return token, nil
	}

	var envelope struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Result != "" {
		return envelope.Result, nil
	}

	trimmed := string(bytes.TrimSpace(body))
	if trimmed != "" {
		return trimmed, nil
	}
	return "", fmt.Errorf("Cloudflare token response was empty")
}

func (c Client) get(ctx context.Context, path string, query map[string]string, apiToken string) ([]byte, error) {
	return c.doJSON(ctx, http.MethodGet, path, query, nil, apiToken)
}

func (c Client) post(ctx context.Context, path string, body any, apiToken string) ([]byte, error) {
	return c.doJSON(ctx, http.MethodPost, path, nil, body, apiToken)
}

func (c Client) put(ctx context.Context, path string, body any, apiToken string) ([]byte, error) {
	return c.doJSON(ctx, http.MethodPut, path, nil, body, apiToken)
}

func (c Client) doJSON(ctx context.Context, method, path string, query map[string]string, body any, apiToken string) ([]byte, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("api token is required")
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	var requestBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	values := req.URL.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	req.URL.RawQuery = values.Encode()

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Cloudflare request failed with %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	return responseBody, nil
}

func (c Client) dnsRecords(ctx context.Context, zoneID, name, recordType, apiToken string) ([]DNSRecord, error) {
	body, err := c.get(ctx, fmt.Sprintf("/zones/%s/dns_records", zoneID), map[string]string{
		"name":     name,
		"type":     recordType,
		"per_page": "10",
	}, apiToken)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Result []DNSRecord `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return envelope.Result, nil
}

func parseDNSRecord(body []byte) (DNSRecord, error) {
	var envelope struct {
		Result DNSRecord `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return DNSRecord{}, err
	}
	return envelope.Result, nil
}

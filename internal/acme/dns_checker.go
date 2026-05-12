package acme

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSCheckResult represents the result of a DNS TXT record check
type DNSCheckResult struct {
	Domain       string   `json:"domain"`
	TXTHost      string   `json:"txt_host"`
	ExpectedValue string  `json:"expected_value"`
	FoundValues  []string `json:"found_values"`
	Matched      bool     `json:"matched"`
	Error        string   `json:"error,omitempty"`
}

// DNSChecker checks DNS TXT records for ACME challenges
type DNSChecker struct {
	resolvers []string
	timeout   time.Duration
}

// NewDNSChecker creates a new DNSChecker with the specified resolvers and timeout.
// resolvers should be in the format "ip:port" (e.g., "8.8.8.8:53")
func NewDNSChecker(resolvers []string, timeout time.Duration) *DNSChecker {
	if len(resolvers) == 0 {
		resolvers = []string{"8.8.8.8:53", "1.1.1.1:53"}
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &DNSChecker{
		resolvers: resolvers,
		timeout:   timeout,
	}
}

// ParseResolvers parses a comma-separated list of DNS resolvers.
func ParseResolvers(resolversStr string) []string {
	if resolversStr == "" {
		return nil
	}
	parts := strings.Split(resolversStr, ",")
	resolvers := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			// Add default port if not specified
			if !strings.Contains(p, ":") {
				p = p + ":53"
			}
			resolvers = append(resolvers, p)
		}
	}
	return resolvers
}

// CheckTXTRecord queries DNS for TXT records and checks if the expected value is present.
func (c *DNSChecker) CheckTXTRecord(ctx context.Context, txtHost, expectedValue string) DNSCheckResult {
	result := DNSCheckResult{
		TXTHost:       txtHost,
		ExpectedValue: expectedValue,
		Matched:       false,
	}

	// Try each resolver until we get a successful response
	var lastErr error
	for _, resolver := range c.resolvers {
		records, err := c.queryTXTRecords(ctx, txtHost, resolver)
		if err != nil {
			lastErr = err
			continue
		}

		result.FoundValues = records
		for _, record := range records {
			if record == expectedValue {
				result.Matched = true
				return result
			}
		}
		// We got records but none matched
		return result
	}

	if lastErr != nil {
		result.Error = lastErr.Error()
	}
	return result
}

// CheckMultipleTXTRecords checks multiple TXT records and returns results for each.
func (c *DNSChecker) CheckMultipleTXTRecords(ctx context.Context, checks []struct {
	Domain        string
	TXTHost       string
	ExpectedValue string
}) []DNSCheckResult {
	results := make([]DNSCheckResult, len(checks))
	for i, check := range checks {
		result := c.CheckTXTRecord(ctx, check.TXTHost, check.ExpectedValue)
		result.Domain = check.Domain
		results[i] = result
	}
	return results
}

// queryTXTRecords queries TXT records using a specific DNS resolver.
func (c *DNSChecker) queryTXTRecords(ctx context.Context, host, resolver string) ([]string, error) {
	// Create a custom resolver
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: c.timeout,
			}
			return d.DialContext(ctx, "udp", resolver)
		},
	}

	// Set context timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Query TXT records
	records, err := r.LookupTXT(ctx, host)
	if err != nil {
		// Check if it's a "no such host" error, which means no TXT records exist
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("DNS lookup failed: %w", err)
	}

	return records, nil
}

// AllMatched returns true if all results have Matched set to true.
func AllMatched(results []DNSCheckResult) bool {
	for _, r := range results {
		if !r.Matched {
			return false
		}
	}
	return len(results) > 0
}

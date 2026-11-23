package utils

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

var blockedHTTPNetworks = func() []netip.Prefix {
	blocks := []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16",
		"172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16",
		"198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
		"::/128", "::1/128", "64:ff9b::/96", "64:ff9b:1::/48", "100::/64", "2001::/23", "2001:db8::/32",
		"2002::/16", "fc00::/7", "fe80::/10", "fec0::/10", "ff00::/8",
	}
	prefixes := make([]netip.Prefix, 0, len(blocks))
	for _, block := range blocks {
		prefixes = append(prefixes, netip.MustParsePrefix(block))
	}
	return prefixes
}()

func isPublicHTTPAddress(ip net.IP) bool {
	addr, err := netip.ParseAddr(ip.String())
	if err != nil {
		return false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	for _, prefix := range blockedHTTPNetworks {
		if prefix.Contains(addr) {
			return false
		}
	}
	return true
}

func publicHTTPDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	if addr, parseErr := netip.ParseAddr(host); parseErr == nil {
		ips = []net.IP{net.ParseIP(addr.String())}
	} else {
		ips, err = net.DefaultResolver.LookupIP(ctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("remote host cannot be resolved")
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("remote host cannot be resolved")
	}
	for _, ip := range ips {
		if !isPublicHTTPAddress(ip) {
			return nil, fmt.Errorf("remote host resolves to a non-public address")
		}
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	var lastErr error
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, lastErr
}

func validatePublicHTTPURL(u *url.URL) error {
	if u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return fmt.Errorf("remote URL must use HTTP or HTTPS without credentials")
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if strings.Contains(host, "%") {
		return fmt.Errorf("scoped IP addresses are not allowed")
	}
	if addr, err := netip.ParseAddr(host); err == nil && !isPublicHTTPAddress(net.ParseIP(addr.String())) {
		return fmt.Errorf("remote URL points to a non-public address")
	}
	return nil
}

// FetchPublicHTTP retrieves a bounded HTTP response from public IP addresses only.
// It resolves and validates each destination at dial time, including redirects.
func FetchPublicHTTP(rawURL string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 || maxBytes == int64(^uint64(0)>>1) {
		return nil, fmt.Errorf("response size limit must be positive")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid remote URL")
	}
	if err := validatePublicHTTPURL(u); err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           publicHTTPDialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return validatePublicHTTPURL(req.URL)
		},
	}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("fetch remote URL failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("remote URL returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read remote response failed")
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("remote response exceeds size limit")
	}
	return body, nil
}

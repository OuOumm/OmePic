package util

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var cgnatNet = mustParseCIDR("100.64.0.0/10")

func mustParseCIDR(value string) *net.IPNet {
	_, parsed, err := net.ParseCIDR(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

// ValidateRemoteImageURL rejects unsupported schemes and malformed remote-image URLs.
func ValidateRemoteImageURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("url must be absolute")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("url credentials are not allowed")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("only http and https urls are allowed")
	}
	return parsed, nil
}

// ValidateResolvedIP rejects local, private, multicast, link-local, unspecified, CGNAT, and other non-global addresses.
func ValidateResolvedIP(ip net.IP) error {
	if ip == nil {
		return fmt.Errorf("ip address is invalid")
	}
	if ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return fmt.Errorf("target address is not allowed")
	}
	if cgnatNet.Contains(ip) {
		return fmt.Errorf("target address is not allowed")
	}
	return nil
}

// ResolveAndValidateHost resolves host and returns public IP candidates. Dialers should still connect to a checked IP.
func ResolveAndValidateHost(ctx context.Context, resolver *net.Resolver, host string) ([]net.IP, error) {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	literal := net.ParseIP(host)
	if literal != nil {
		if err := ValidateResolvedIP(literal); err != nil {
			return nil, err
		}
		return []net.IP{literal}, nil
	}
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target host")
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if err := ValidateResolvedIP(addr.IP); err != nil {
			return nil, err
		}
		ips = append(ips, addr.IP)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("target host has no addresses")
	}
	return ips, nil
}

package clientip

import (
	"net"
	"net/http"
	"strings"
)

type Resolver struct {
	trustedProxyCIDRs []*net.IPNet
	realIPHeader      string
}

func NewResolver(trustedProxyCIDRs []string, realIPHeader string) *Resolver {
	resolver := &Resolver{realIPHeader: normalizeHeader(realIPHeader)}
	for _, value := range trustedProxyCIDRs {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, cidr, err := net.ParseCIDR(trimmed); err == nil {
			resolver.trustedProxyCIDRs = append(resolver.trustedProxyCIDRs, cidr)
			continue
		}
		if ip := net.ParseIP(trimmed); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			resolver.trustedProxyCIDRs = append(resolver.trustedProxyCIDRs, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		}
	}
	return resolver
}

func (r *Resolver) Resolve(req *http.Request) string {
	if req == nil {
		return ""
	}
	remoteIP := parseRemoteIP(req.RemoteAddr)
	if remoteIP == "" {
		return ""
	}
	if !r.isTrustedProxy(remoteIP) {
		return remoteIP
	}
	if headerIP := r.headerIP(req); headerIP != "" {
		return headerIP
	}
	return remoteIP
}

func (r *Resolver) isTrustedProxy(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil || len(r.trustedProxyCIDRs) == 0 {
		return false
	}
	for _, cidr := range r.trustedProxyCIDRs {
		if cidr.Contains(parsed) {
			return true
		}
	}
	return false
}

func (r *Resolver) headerIP(req *http.Request) string {
	switch r.realIPHeader {
	case "x-real-ip":
		return firstValidIP(req.Header.Get("X-Real-IP"))
	default:
		return firstForwardedIP(req.Header.Get("X-Forwarded-For"))
	}
}

func parseRemoteIP(remoteAddr string) string {
	trimmed := strings.TrimSpace(remoteAddr)
	if trimmed == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		trimmed = host
	}
	return firstValidIP(trimmed)
}

func firstForwardedIP(value string) string {
	for _, part := range strings.Split(value, ",") {
		if ip := firstValidIP(part); ip != "" {
			return ip
		}
	}
	return ""
}

func firstValidIP(value string) string {
	parsed := net.ParseIP(strings.TrimSpace(value))
	if parsed == nil {
		return ""
	}
	return parsed.String()
}

func normalizeHeader(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "x-real-ip" {
		return normalized
	}
	return "x-forwarded-for"
}

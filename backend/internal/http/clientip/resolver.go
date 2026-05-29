package clientip

import (
	"net"
	"net/http"
	"strings"
)

// Resolver extracts the real client IP from an HTTP request.
//
// The header source is resolved dynamically on every call via the
// realIPSourceFunc callback so that runtime settings changes (hot-reload)
// take effect without restarting the server.
type Resolver struct {
	trustedProxyCIDRs []*net.IPNet
	realIPSourceFunc  func() string
}

// NewResolver creates a Resolver.
//
// trustedProxyCIDRs is a list of CIDR ranges or single IPs that are
// considered trusted proxies; only requests from these addresses will
// have their forwarding headers honoured.
//
// realIPSourceFunc is called on every Resolve invocation to determine
// which header (or direct RemoteAddr) to use.  If nil, the resolver
// always returns RemoteAddr.  Recognised return values:
//
//   - "remote-addr"     – use RemoteAddr directly (default, no proxy)
//   - "x-forwarded-for" – first IP from X-Forwarded-For
//   - "x-real-ip"       – X-Real-IP header
//   - "cf-connecting-ip"– CF-Connecting-IP header (Cloudflare)
func NewResolver(trustedProxyCIDRs []string, realIPSourceFunc func() string) *Resolver {
	resolver := &Resolver{realIPSourceFunc: realIPSourceFunc}
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

	source := ""
	if r.realIPSourceFunc != nil {
		source = r.realIPSourceFunc()
	}

	// remote-addr or empty → always use direct IP, ignore headers.
	if source == "" || source == "remote-addr" {
		return remoteIP
	}

	// For header-based sources, only honour the header when the direct
	// connection comes from a trusted proxy.
	if !r.isTrustedProxy(remoteIP) {
		return remoteIP
	}
	if headerIP := r.headerIP(req, source); headerIP != "" {
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

func (r *Resolver) headerIP(req *http.Request, source string) string {
	switch source {
	case "x-real-ip":
		return firstValidIP(req.Header.Get("X-Real-IP"))
	case "cf-connecting-ip":
		return firstValidIP(req.Header.Get("CF-Connecting-IP"))
	case "x-forwarded-for":
		return firstForwardedIP(req.Header.Get("X-Forwarded-For"))
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

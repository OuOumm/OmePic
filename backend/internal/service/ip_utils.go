package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

func ipHash(ipAddress string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(ipAddress)))
	return hex.EncodeToString(sum[:])
}

func maskIPAddress(ipAddress string) string {
	parsed := net.ParseIP(strings.TrimSpace(ipAddress))
	if parsed == nil {
		return ""
	}
	if v4 := parsed.To4(); v4 != nil {
		return strings.Join([]string{strconv.Itoa(int(v4[0])), strconv.Itoa(int(v4[1])), strconv.Itoa(int(v4[2])), "*"}, ".")
	}
	parts := strings.Split(parsed.String(), ":")
	if len(parts) <= 2 {
		return parsed.String()
	}
	return strings.Join(parts[:2], ":") + ":*"
}

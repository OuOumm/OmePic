package util

import (
	"context"
	"net"
	"testing"
)

func TestValidateRemoteImageURLRejectsInvalidSchemes(t *testing.T) {
	for _, rawURL := range []string{"ftp://example.com/a.png", "file:///etc/passwd", "//example.com/a.png", "http://user:pass@example.com/a.png"} {
		if _, err := ValidateRemoteImageURL(rawURL); err == nil {
			t.Fatalf("expected %q to be rejected", rawURL)
		}
	}
	if _, err := ValidateRemoteImageURL("https://example.com/a.png"); err != nil {
		t.Fatalf("expected https url to pass: %v", err)
	}
}

func TestValidateResolvedIPRejectsUnsafeRanges(t *testing.T) {
	unsafe := []string{
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"127.0.0.1",
		"169.254.169.254",
		"224.0.0.1",
		"0.0.0.0",
		"100.64.0.1",
		"::1",
		"fe80::1",
		"ff02::1",
		"::",
	}
	for _, value := range unsafe {
		if err := ValidateResolvedIP(net.ParseIP(value)); err == nil {
			t.Fatalf("expected %s to be rejected", value)
		}
	}
	for _, value := range []string{"8.8.8.8", "2001:4860:4860::8888"} {
		if err := ValidateResolvedIP(net.ParseIP(value)); err != nil {
			t.Fatalf("expected %s to pass: %v", value, err)
		}
	}
}

func TestResolveAndValidateHostRejectsLiteralPrivateIP(t *testing.T) {
	if _, err := ResolveAndValidateHost(context.Background(), nil, "127.0.0.1"); err == nil {
		t.Fatal("expected loopback literal to be rejected")
	}
}

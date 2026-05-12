package iputil

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHashTrimsInputAndPreservesAlgorithm(t *testing.T) {
	wantSum := sha256.Sum256([]byte("203.0.113.8"))
	want := hex.EncodeToString(wantSum[:])

	if got := Hash(" 203.0.113.8 \n"); got != want {
		t.Fatalf("unexpected hash: got %q want %q", got, want)
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{name: "empty", ip: "", want: ""},
		{name: "invalid", ip: "not-an-ip", want: ""},
		{name: "ipv4", ip: " 203.0.113.8 ", want: "203.0.113.*"},
		{name: "ipv6", ip: "2001:db8:85a3::8a2e:370:7334", want: "2001:db8:*"},
		{name: "ipv4 mapped ipv6", ip: "::ffff:192.0.2.128", want: "192.0.2.*"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Mask(test.ip); got != test.want {
				t.Fatalf("unexpected mask: got %q want %q", got, test.want)
			}
		})
	}
}

package auth

import "testing"

func TestParseBearer(t *testing.T) {
	token, err := ParseBearer("Bearer abc")
	if err != nil {
		t.Fatalf("ParseBearer returned error: %v", err)
	}
	if token != "abc" {
		t.Fatalf("expected token abc, got %s", token)
	}
}

package secrets

import (
	"testing"
)

func TestSealOpenRoundtrip(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz012345") // 32 bytes
	plaintext := "my-s3-secret-access-key-12345"

	ciphertext, err := Seal(plaintext, key)
	if err != nil {
		t.Fatalf("Seal returned error: %v", err)
	}
	if ciphertext == plaintext {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := Open(ciphertext, key)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestSealOpenRoundtripEmpty(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz012345")

	ciphertext, err := Seal("", key)
	if err != nil {
		t.Fatalf("Seal of empty string returned error: %v", err)
	}

	decrypted, err := Open(ciphertext, key)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if decrypted != "" {
		t.Errorf("decrypted = %q, want empty string", decrypted)
	}
}

func TestSealDifferentNonces(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz012345")
	plaintext := "same-input"

	ct1, err := Seal(plaintext, key)
	if err != nil {
		t.Fatalf("Seal returned error: %v", err)
	}
	ct2, err := Seal(plaintext, key)
	if err != nil {
		t.Fatalf("Seal returned error: %v", err)
	}
	if ct1 == ct2 {
		t.Error("two Seal calls with same plaintext should produce different ciphertexts (different nonce)")
	}
}

func TestOpenWrongKey(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz012345")
	wrongKey := []byte("012345abcdefghijklmnopqrstuvwxyz") // also 32 bytes, different

	ciphertext, err := Seal("secret-value", key)
	if err != nil {
		t.Fatalf("Seal returned error: %v", err)
	}

	_, err = Open(ciphertext, wrongKey)
	if err == nil {
		t.Error("Open with wrong key should return an error")
	}
}

func TestOpenInvalidBase64(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz012345")

	_, err := Open("not-valid-base64!!!", key)
	if err == nil {
		t.Error("Open with invalid base64 should return an error")
	}
}

func TestOpenTooShort(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz012345")

	// base64 of just a few bytes — shorter than nonce size (12)
	shortCT := "YWJj" // base64("abc") = 3 bytes

	_, err := Open(shortCT, key)
	if err == nil {
		t.Error("Open with ciphertext shorter than nonce should return an error")
	}
	if err.Error() != "ciphertext too short" {
		t.Errorf("error = %q, want %q", err.Error(), "ciphertext too short")
	}
}

func TestSealInvalidKeyTooShort(t *testing.T) {
	_, err := Seal("plaintext", []byte("short"))
	if err != ErrInvalidKey {
		t.Errorf("error = %v, want ErrInvalidKey", err)
	}
}

func TestSealInvalidKeyTooLong(t *testing.T) {
	_, err := Seal("plaintext", []byte("this-key-is-way-too-long-for-aes-256-gcm"))
	if err != ErrInvalidKey {
		t.Errorf("error = %v, want ErrInvalidKey", err)
	}
}

func TestOpenInvalidKeyTooShort(t *testing.T) {
	_, err := Open("YWJj", []byte("short"))
	if err != ErrInvalidKey {
		t.Errorf("error = %v, want ErrInvalidKey", err)
	}
}

func TestOpenInvalidKeyTooLong(t *testing.T) {
	_, err := Open("YWJj", []byte("this-key-is-way-too-long-for-aes-256-gcm"))
	if err != ErrInvalidKey {
		t.Errorf("error = %v, want ErrInvalidKey", err)
	}
}

func TestEncryptorRoundtrip(t *testing.T) {
	enc, err := NewEncryptor("abcdefghijklmnopqrstuvwxyz012345")
	if err != nil {
		t.Fatalf("NewEncryptor returned error: %v", err)
	}

	plaintext := "AKIAIOSFODNN7EXAMPLE"
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestNewEncryptorInvalidKey(t *testing.T) {
	_, err := NewEncryptor("short-key")
	if err != ErrInvalidKey {
		t.Errorf("error = %v, want ErrInvalidKey", err)
	}
}

func TestEncryptorEncryptDecryptMultipleFields(t *testing.T) {
	enc, err := NewEncryptor("abcdefghijklmnopqrstuvwxyz012345")
	if err != nil {
		t.Fatalf("NewEncryptor returned error: %v", err)
	}

	fields := []string{
		"s3-access-key-value",
		"s3-secret-key-value-with-special-chars!@#$",
		"webdav-password-12345",
		"", // empty field
	}

	for _, field := range fields {
		ct, err := enc.Encrypt(field)
		if err != nil {
			t.Fatalf("Encrypt(%q) returned error: %v", field, err)
		}
		pt, err := enc.Decrypt(ct)
		if err != nil {
			t.Fatalf("Decrypt returned error: %v", err)
		}
		if pt != field {
			t.Errorf("field %q: decrypted = %q, want %q", field, pt, field)
		}
	}
}
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// ErrInvalidKey is returned when the encryption key is not 32 bytes.
var ErrInvalidKey = errors.New("secret encryption key must be exactly 32 bytes")

// Seal encrypts plaintext using AES-256-GCM and returns a base64-encoded
// ciphertext string containing nonce+ciphertext+tag. Format: base64(nonce[12] + ciphertext+tag)
func Seal(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a base64-encoded AES-256-GCM ciphertext back to plaintext.
func Open(ciphertext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKey
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// SecretEncryptor holds the key and provides convenience methods for
// encrypting/decrypting storage credential fields.
type SecretEncryptor struct {
	key []byte
}

// NewEncryptor creates a SecretEncryptor from a raw key string.
// The key must be exactly 32 bytes (256 bits) for AES-256-GCM.
func NewEncryptor(keyString string) (*SecretEncryptor, error) {
	key := []byte(keyString)
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return &SecretEncryptor{key: key}, nil
}

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext.
func (e *SecretEncryptor) Encrypt(plaintext string) (string, error) {
	return Seal(plaintext, e.key)
}

// Decrypt decrypts a base64-encoded ciphertext back to plaintext.
func (e *SecretEncryptor) Decrypt(ciphertext string) (string, error) {
	return Open(ciphertext, e.key)
}
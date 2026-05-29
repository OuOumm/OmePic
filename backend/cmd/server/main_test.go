package main

import (
	"testing"

	"omepic/backend/internal/config"
)

func TestEnforceRequiredSecrets(t *testing.T) {
	// 32-char valid secret for positive cases
	valid32 := "abcdefghijklmnopqrstuvwxyz123456" // 32 chars

	tests := []struct {
		name    string
		cfg     config.AppConfig
		wantErr bool
		errMsg  string // substring expected in error message
	}{
		{
			name: "JWT_SECRET missing",
			cfg:  config.AppConfig{JWTSecret: "", UIDEncryptionKey: valid32, SecretEncryptionKey: valid32},
			wantErr: true,
			errMsg: "JWT_SECRET",
		},
		{
			name: "JWT_SECRET too short",
			cfg:  config.AppConfig{JWTSecret: "short", UIDEncryptionKey: valid32, SecretEncryptionKey: valid32},
			wantErr: true,
			errMsg: "JWT_SECRET",
		},
		{
			name: "UID_ENCRYPTION_KEY missing",
			cfg:  config.AppConfig{JWTSecret: valid32, UIDEncryptionKey: "", SecretEncryptionKey: valid32},
			wantErr: true,
			errMsg: "UID_ENCRYPTION_KEY",
		},
		{
			name: "UID_ENCRYPTION_KEY too short",
			cfg:  config.AppConfig{JWTSecret: valid32, UIDEncryptionKey: "short", SecretEncryptionKey: valid32},
			wantErr: true,
			errMsg: "UID_ENCRYPTION_KEY",
		},
		{
			name: "SECRET_ENCRYPTION_KEY missing",
			cfg:  config.AppConfig{JWTSecret: valid32, UIDEncryptionKey: valid32, SecretEncryptionKey: ""},
			wantErr: true,
			errMsg: "SECRET_ENCRYPTION_KEY",
		},
		{
			name: "SECRET_ENCRYPTION_KEY too short",
			cfg:  config.AppConfig{JWTSecret: valid32, UIDEncryptionKey: valid32, SecretEncryptionKey: "short"},
			wantErr: true,
			errMsg: "SECRET_ENCRYPTION_KEY",
		},
		{
			name:    "all secrets valid",
			cfg:     config.AppConfig{JWTSecret: valid32, UIDEncryptionKey: valid32, SecretEncryptionKey: valid32},
			wantErr: false,
		},
		{
			name:    "JWT_SECRET exactly 32 chars passes",
			cfg:     config.AppConfig{JWTSecret: "abcdefghijklmnopqrstuvwxyz123456", UIDEncryptionKey: valid32, SecretEncryptionKey: valid32},
			wantErr: false,
		},
		{
			name:    "JWT_SECRET 31 chars fails",
			cfg:     config.AppConfig{JWTSecret: "abcdefghijklmnopqrstuvwxyz12345", UIDEncryptionKey: valid32, SecretEncryptionKey: valid32},
			wantErr: true,
			errMsg: "JWT_SECRET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforceRequiredSecrets(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Fatalf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && len(sub) > 0 && findSubstring(s, sub)))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
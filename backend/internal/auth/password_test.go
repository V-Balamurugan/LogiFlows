package auth_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/auth"
)

func TestValidatePassword_Rules(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "SecureP@ssw0rd!", false},
		{"Too short", "Sh0rt!", true},
		{"Missing uppercase", "securep@ssw0rd!", true},
		{"Missing lowercase", "SECUREP@SSW0RD!", true},
		{"Missing digit", "SecureP@ssword!", true},
		{"Missing symbol", "SecurePassw0rd1", true},
		{"Empty password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashAndComparePassword(t *testing.T) {
	plain := "ValidSecure#2026Pass"

	hash, err := auth.HashPassword(plain)
	if err != nil {
		t.Fatalf("expected HashPassword to succeed, got %v", err)
	}

	if hash == plain {
		t.Errorf("hash must not equal plain text password")
	}

	if !auth.ComparePassword(hash, plain) {
		t.Errorf("expected ComparePassword to return true for matching password")
	}

	if auth.ComparePassword(hash, "WrongPassword123!") {
		t.Errorf("expected ComparePassword to return false for non-matching password")
	}
}

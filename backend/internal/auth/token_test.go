package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/auth"
)

func TestTokenService_GenerateAndValidate(t *testing.T) {
	secret := "a-very-long-and-secure-test-jwt-secret-key-32chars"
	service := auth.NewTokenService(secret, 15*time.Minute, "logiflows-test")

	userID := uuid.New()
	email := "operator@quickcargo.com"

	token, expiresAt, err := service.GenerateAccessToken(userID, email, false)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	if expiresAt.Before(time.Now()) {
		t.Errorf("expected expiresAt to be in the future, got %v", expiresAt)
	}

	claims, err := service.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user_id %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.IsPlatformAdmin {
		t.Errorf("expected IsPlatformAdmin to be false")
	}
}

func TestTokenService_ExpiredToken(t *testing.T) {
	secret := "a-very-long-and-secure-test-jwt-secret-key-32chars"
	// Create token that expires in -1 second (already expired)
	service := auth.NewTokenService(secret, -1*time.Second, "logiflows-test")

	token, _, err := service.GenerateAccessToken(uuid.New(), "test@logiflows.com", false)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = service.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error validating expired token, got nil")
	}
	if !errors.Is(err, auth.ErrExpiredToken) {
		t.Errorf("expected ErrExpiredToken, got: %v", err)
	}
}

func TestTokenService_WrongSecret(t *testing.T) {
	secretA := "secret-key-number-one-with-more-than-32-chars!"
	secretB := "secret-key-number-two-with-more-than-32-chars!"

	serviceA := auth.NewTokenService(secretA, 15*time.Minute, "logiflows-test")
	serviceB := auth.NewTokenService(secretB, 15*time.Minute, "logiflows-test")

	token, _, err := serviceA.GenerateAccessToken(uuid.New(), "test@logiflows.com", false)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = serviceB.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected validation failure with wrong secret, got nil")
	}
}

func TestGenerateRefreshToken_FormatAndUniqueness(t *testing.T) {
	rawToken1, hash1, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("expected GenerateRefreshToken to succeed, got %v", err)
	}

	if len(rawToken1) != 64 {
		t.Errorf("expected 64-char hex string (32 bytes), got length %d", len(rawToken1))
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char SHA-256 hex digest, got length %d", len(hash1))
	}

	// Verify hash function matches
	computedHash := auth.HashRefreshToken(rawToken1)
	if computedHash != hash1 {
		t.Errorf("expected HashRefreshToken to match generated hash: got %s, want %s", computedHash, hash1)
	}

	// Verify uniqueness across consecutive calls
	rawToken2, hash2, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("expected second GenerateRefreshToken to succeed, got %v", err)
	}

	if rawToken1 == rawToken2 {
		t.Errorf("expected distinct cryptographically random tokens, got identical: %s", rawToken1)
	}
	if hash1 == hash2 {
		t.Errorf("expected distinct hashes for distinct tokens, got identical: %s", hash1)
	}
}

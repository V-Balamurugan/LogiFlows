package users_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/users"
)

func TestUser_PasswordHash_ExcludedFromJSON(t *testing.T) {
	phone := "+1-555-0199"
	now := time.Now().UTC()
	user := users.User{
		ID:              uuid.New(),
		Email:           "secure.user@logiflows.com",
		PasswordHash:    "$2a$12$SuperSecretHashedPasswordNeverLeakThisValue",
		FullName:        "Secure User",
		PhoneNumber:     &phone,
		IsActive:        true,
		IsPlatformAdmin: false,
		EmailVerified:   true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	bytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	jsonStr := string(bytes)

	// Critical security validation: password_hash must NEVER appear in serialized JSON
	if strings.Contains(jsonStr, "password_hash") {
		t.Errorf("SECURITY LEAK: JSON serialization contains 'password_hash' key: %s", jsonStr)
	}
	if strings.Contains(jsonStr, "SuperSecretHashedPasswordNeverLeakThisValue") {
		t.Errorf("SECURITY LEAK: JSON serialization leaked password hash value: %s", jsonStr)
	}

	// Verify required fields are present
	if !strings.Contains(jsonStr, "email_verified") {
		t.Errorf("expected JSON to include 'email_verified': %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "secure.user@logiflows.com") {
		t.Errorf("expected JSON to include email: %s", jsonStr)
	}
}

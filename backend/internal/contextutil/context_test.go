package contextutil_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/contextutil"
)

func TestContextUtil_UserAccessors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Initially empty
	if id, ok := contextutil.GetUserID(c); ok || id != uuid.Nil {
		t.Errorf("expected empty user ID, got %s", id)
	}
	if email := contextutil.GetUserEmail(c); email != "" {
		t.Errorf("expected empty email, got %s", email)
	}
	if isAdmin := contextutil.IsPlatformAdmin(c); isAdmin {
		t.Errorf("expected isPlatformAdmin to be false")
	}

	// Set and verify
	expectedID := uuid.New()
	expectedEmail := "admin@logiflows.com"

	contextutil.SetUserID(c, expectedID)
	contextutil.SetUserEmail(c, expectedEmail)
	contextutil.SetIsPlatformAdmin(c, true)

	if id, ok := contextutil.GetUserID(c); !ok || id != expectedID {
		t.Errorf("expected user ID %s, got %s", expectedID, id)
	}
	if email := contextutil.GetUserEmail(c); email != expectedEmail {
		t.Errorf("expected email %s, got %s", expectedEmail, email)
	}
	if isAdmin := contextutil.IsPlatformAdmin(c); !isAdmin {
		t.Errorf("expected isPlatformAdmin to be true")
	}
}

func TestContextUtil_TenantAccessors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Initially empty
	if id, ok := contextutil.GetTenantID(c); ok || id != uuid.Nil {
		t.Errorf("expected empty tenant ID, got %s", id)
	}
	if role := contextutil.GetTenantRole(c); role != "" {
		t.Errorf("expected empty tenant role, got %s", role)
	}

	// Set and verify
	expectedTenantID := uuid.New()
	expectedRole := "TENANT_ADMIN"

	contextutil.SetTenantID(c, expectedTenantID)
	contextutil.SetTenantRole(c, expectedRole)

	if id, ok := contextutil.GetTenantID(c); !ok || id != expectedTenantID {
		t.Errorf("expected tenant ID %s, got %s", expectedTenantID, id)
	}
	if role := contextutil.GetTenantRole(c); role != expectedRole {
		t.Errorf("expected role %s, got %s", expectedRole, role)
	}
}

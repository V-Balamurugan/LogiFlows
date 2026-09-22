package authorization_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/authorization"
)

func TestRolePermissions(t *testing.T) {
	tests := []struct {
		role            string
		isPlatformAdmin bool
		canManage       bool
		canOperate      bool
		canView         bool
	}{
		// Platform admin can do everything
		{authorization.RolePlatformAdmin, false, true, true, true},
		{authorization.RoleViewer, true, true, true, true},

		// Tenant admin can manage, operate, and view
		{authorization.RoleTenantAdmin, false, true, true, true},

		// Tenant operator can operate and view, but not manage
		{authorization.RoleTenantOperator, false, false, true, true},

		// Viewer can only view
		{authorization.RoleViewer, false, false, false, true},

		// Unknown role can do nothing
		{"UNKNOWN_ROLE", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if got := authorization.CanManageTenant(tt.role, tt.isPlatformAdmin); got != tt.canManage {
				t.Errorf("CanManageTenant(%s, %v) = %v; want %v", tt.role, tt.isPlatformAdmin, got, tt.canManage)
			}
			if got := authorization.CanOperateTenant(tt.role, tt.isPlatformAdmin); got != tt.canOperate {
				t.Errorf("CanOperateTenant(%s, %v) = %v; want %v", tt.role, tt.isPlatformAdmin, got, tt.canOperate)
			}
			if got := authorization.CanViewTenant(tt.role, tt.isPlatformAdmin); got != tt.canView {
				t.Errorf("CanViewTenant(%s, %v) = %v; want %v", tt.role, tt.isPlatformAdmin, got, tt.canView)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	for _, r := range authorization.ValidRoles {
		if !authorization.IsValidRole(r) {
			t.Errorf("expected %s to be valid role", r)
		}
	}
	if authorization.IsValidRole("INVALID_ROLE") {
		t.Errorf("expected INVALID_ROLE to be false")
	}
}

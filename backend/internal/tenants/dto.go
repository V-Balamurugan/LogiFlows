package tenants

// CreateTenantRequest holds the input for creating a new tenant company.
type CreateTenantRequest struct {
	Name         string `json:"name" binding:"required"`
	ContactEmail string `json:"contact_email" binding:"required,email"`
}

// AddMemberRequest holds the input for adding a member to a tenant.
type AddMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// UpdateTenantRequest holds the input for updating tenant metadata.
type UpdateTenantRequest struct {
	Name         *string `json:"name"`
	ContactEmail *string `json:"contact_email" binding:"omitempty,email"`
}

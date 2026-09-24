package customers_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/customers"
)

func TestCustomer_ValidationRules(t *testing.T) {
	validReq := customers.CreateCustomerRequest{
		Name:         "Acme Corporation",
		CustomerType: customers.TypeBusiness,
		Email:        ptr("billing@acme.com"),
		Phone:        "+91 9876543210",
		AddressLine1: "123 Commercial St",
		City:         "Chennai",
		State:        "Tamil Nadu",
		PostalCode:   "600001",
		Country:      "India",
	}

	if err := customers.ValidateCreateRequest(validReq); err != nil {
		t.Fatalf("expected valid customer request to pass, got: %v", err)
	}

	// Test missing name
	invalidReq := validReq
	invalidReq.Name = ""
	if err := customers.ValidateCreateRequest(invalidReq); err != customers.ErrInvalidCustomerName {
		t.Errorf("expected ErrInvalidCustomerName, got: %v", err)
	}

	// Test missing phone
	invalidReq = validReq
	invalidReq.Phone = "   "
	if err := customers.ValidateCreateRequest(invalidReq); err != customers.ErrInvalidCustomerPhone {
		t.Errorf("expected ErrInvalidCustomerPhone, got: %v", err)
	}

	// Test missing address line 1
	invalidReq = validReq
	invalidReq.AddressLine1 = ""
	if err := customers.ValidateCreateRequest(invalidReq); err != customers.ErrInvalidCustomerAddress {
		t.Errorf("expected ErrInvalidCustomerAddress, got: %v", err)
	}

	// Test missing city
	invalidReq = validReq
	invalidReq.City = ""
	if err := customers.ValidateCreateRequest(invalidReq); err != customers.ErrInvalidCustomerAddress {
		t.Errorf("expected ErrInvalidCustomerAddress, got: %v", err)
	}

	// Test missing postal code
	invalidReq = validReq
	invalidReq.PostalCode = ""
	if err := customers.ValidateCreateRequest(invalidReq); err != customers.ErrInvalidCustomerAddress {
		t.Errorf("expected ErrInvalidCustomerAddress, got: %v", err)
	}

	// Test invalid customer type
	invalidReq = validReq
	invalidReq.CustomerType = "UNKNOWN_VIP_TYPE"
	if err := customers.ValidateCreateRequest(invalidReq); err != customers.ErrInvalidCustomerType {
		t.Errorf("expected ErrInvalidCustomerType, got: %v", err)
	}
}

func ptr(s string) *string {
	return &s
}

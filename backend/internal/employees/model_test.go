package employees_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/employees"
)

func TestCreateEmployeeRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     employees.CreateEmployeeRequest
		wantErr error
	}{
		{
			name: "Valid Driver With Generated Code Omitted",
			req: employees.CreateEmployeeRequest{
				EmployeeCode:    "",
				FirstName:       "Rajesh",
				LastName:        "Kumar",
				Designation:     "Senior Heavy Vehicle Driver",
				EmploymentType:  "FULL_TIME",
				OperationalRole: "DRIVER",
			},
			wantErr: nil,
		},
		{
			name: "Valid Driver With Explicit Code",
			req: employees.CreateEmployeeRequest{
				EmployeeCode:    "DRV-001",
				FirstName:       "Rajesh",
				LastName:        "Kumar",
				Designation:     "Senior Heavy Vehicle Driver",
				EmploymentType:  "FULL_TIME",
				OperationalRole: "DRIVER",
			},
			wantErr: nil,
		},
		{
			name: "Invalid Code (Too Short)",
			req: employees.CreateEmployeeRequest{
				EmployeeCode:    "D",
				FirstName:       "Rajesh",
				LastName:        "Kumar",
				Designation:     "Driver",
				EmploymentType:  "FULL_TIME",
				OperationalRole: "DRIVER",
			},
			wantErr: employees.ErrInvalidEmployeeCode,
		},
		{
			name: "Missing First Name",
			req: employees.CreateEmployeeRequest{
				EmployeeCode:    "DRV-002",
				FirstName:       "",
				LastName:        "Kumar",
				Designation:     "Driver",
				EmploymentType:  "FULL_TIME",
				OperationalRole: "DRIVER",
			},
			wantErr: employees.ErrInvalidEmployeeName,
		},
		{
			name: "Invalid Employment Type",
			req: employees.CreateEmployeeRequest{
				EmployeeCode:    "DRV-003",
				FirstName:       "Sunil",
				LastName:        "Sharma",
				Designation:     "Driver",
				EmploymentType:  "FREELANCE_INVALID",
				OperationalRole: "DRIVER",
			},
			wantErr: employees.ErrInvalidEmploymentType,
		},
		{
			name: "Invalid Operational Role",
			req: employees.CreateEmployeeRequest{
				EmployeeCode:    "DRV-004",
				FirstName:       "Sunil",
				LastName:        "Sharma",
				Designation:     "Driver",
				EmploymentType:  "FULL_TIME",
				OperationalRole: "CEO_INVALID",
			},
			wantErr: employees.ErrInvalidOperationalRole,
		},
		{
			name: "Invalid Availability Status",
			req: employees.CreateEmployeeRequest{
				FirstName:          "Sunil",
				LastName:           "Sharma",
				Designation:        "Driver",
				AvailabilityStatus: "SLEEPING_INVALID",
			},
			wantErr: employees.ErrInvalidAvailabilityStatus,
		},
		{
			name: "Invalid Verification Status",
			req: employees.CreateEmployeeRequest{
				FirstName:          "Sunil",
				LastName:           "Sharma",
				Designation:        "Driver",
				VerificationStatus: "UNKNOWN_INVALID",
			},
			wantErr: employees.ErrInvalidVerificationStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.ValidateAndSanitize()
			if tt.wantErr != nil && err != tt.wantErr {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestUpdateEmployeeStatusRequest_Validation(t *testing.T) {
	active := employees.StatusActive
	invalidStatus := "NOT_A_STATUS"
	avail := employees.AvailabilityStatusAvailable
	invalidAvail := "NOT_AN_AVAIL"
	verified := employees.VerificationStatusVerified
	invalidVerified := "NOT_VERIFIED"

	tests := []struct {
		name    string
		req     employees.UpdateEmployeeStatusRequest
		wantErr bool
	}{
		{
			name: "Valid Status Only",
			req: employees.UpdateEmployeeStatusRequest{
				Status: &active,
			},
			wantErr: false,
		},
		{
			name: "Valid Status and Availability",
			req: employees.UpdateEmployeeStatusRequest{
				Status:             &active,
				AvailabilityStatus: &avail,
				VerificationStatus: &verified,
			},
			wantErr: false,
		},
		{
			name: "Invalid Status Value",
			req: employees.UpdateEmployeeStatusRequest{
				Status: &invalidStatus,
			},
			wantErr: true,
		},
		{
			name: "Invalid Availability Value",
			req: employees.UpdateEmployeeStatusRequest{
				AvailabilityStatus: &invalidAvail,
			},
			wantErr: true,
		},
		{
			name: "Invalid Verification Value",
			req: employees.UpdateEmployeeStatusRequest{
				VerificationStatus: &invalidVerified,
			},
			wantErr: true,
		},
		{
			name:    "Empty Payload",
			req:     employees.UpdateEmployeeStatusRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.ValidateAndSanitize()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error? %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

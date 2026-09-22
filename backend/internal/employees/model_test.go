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
			name: "Valid Driver",
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

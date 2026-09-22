package vehicles_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/vehicles"
)

func TestCreateVehicleRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     vehicles.CreateVehicleRequest
		wantErr error
	}{
		{
			name: "Valid Electric Van",
			req: vehicles.CreateVehicleRequest{
				RegistrationNumber: "DL-01-EV-4091",
				VehicleType:        "ELECTRIC_VAN",
				MaxWeightKG:        750,
				MaxVolumeCBM:       4.5,
			},
			wantErr: nil,
		},
		{
			name: "Invalid Registration Number",
			req: vehicles.CreateVehicleRequest{
				RegistrationNumber: "X",
				VehicleType:        "ELECTRIC_VAN",
			},
			wantErr: vehicles.ErrInvalidRegistrationNumber,
		},
		{
			name: "Invalid Vehicle Type",
			req: vehicles.CreateVehicleRequest{
				RegistrationNumber: "DL-01-AB-1234",
				VehicleType:        "HELICOPTER_INVALID",
			},
			wantErr: vehicles.ErrInvalidVehicleType,
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

package branches

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateBranchRequest_Validate_Success(t *testing.T) {
	req := CreateBranchRequest{
		BranchCode:       "  brn-delhi-01  ",
		Name:             "  Delhi Central Hub  ",
		Address:          "Plot 42, Connaught Place",
		City:             "New Delhi",
		State:            "Delhi",
		PostalCode:       "110001",
		Country:          "",
		Latitude:         28.6139,
		Longitude:        77.2090,
		CoverageRadiusKM: 20.0,
	}

	err := req.Validate()
	if err != nil {
		t.Fatalf("expected validation success, got: %v", err)
	}

	if req.BranchCode != "BRN-DELHI-01" {
		t.Errorf("expected branch code normalized to uppercase 'BRN-DELHI-01', got: %s", req.BranchCode)
	}
	if req.Country != "India" {
		t.Errorf("expected default country 'India', got: %s", req.Country)
	}
	if req.CoverageRadiusKM != 20.0 {
		t.Errorf("expected coverage radius 20.0, got: %f", req.CoverageRadiusKM)
	}
}

func TestCreateBranchRequest_Validate_DefaultRadius(t *testing.T) {
	req := CreateBranchRequest{
		BranchCode: "HUB-MUM",
		Name:       "Mumbai Port Hub",
		Address:    "Docks Road",
		City:       "Mumbai",
		Latitude:   18.9220,
		Longitude:  72.8347,
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("expected validation success, got: %v", err)
	}

	if req.CoverageRadiusKM != DefaultCoverageRadiusKM {
		t.Errorf("expected default radius %f, got: %f", DefaultCoverageRadiusKM, req.CoverageRadiusKM)
	}
}

func TestCreateBranchRequest_Validate_InvalidCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{"Latitude too high", 91.0, 77.0},
		{"Latitude too low", -91.0, 77.0},
		{"Longitude too high", 28.0, 181.0},
		{"Longitude too low", 28.0, -181.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateBranchRequest{
				BranchCode: "TEST-01",
				Name:       "Test Branch",
				Address:    "Test Address",
				City:       "Test City",
				Latitude:   tc.latitude,
				Longitude:  tc.longitude,
			}
			err := req.Validate()
			if err != ErrInvalidCoordinates {
				t.Errorf("expected ErrInvalidCoordinates, got: %v", err)
			}
		})
	}
}

func TestBranch_JSONSerialization(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	dist := 12.34
	now := time.Now().UTC()

	b := Branch{
		ID:               id,
		TenantID:         tenantID,
		BranchCode:       "BLR-HUB-01",
		Name:             "Bangalore Tech Hub",
		Address:          "Outer Ring Road",
		City:             "Bangalore",
		State:            "Karnataka",
		PostalCode:       "560103",
		Country:          "India",
		Latitude:         12.9716,
		Longitude:        77.5946,
		CoverageRadiusKM: 15.0,
		Status:           StatusActive,
		IsActive:         true,
		DistanceKM:       &dist,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	bytes, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("failed to marshal branch: %v", err)
	}

	var parsed Branch
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal branch: %v", err)
	}

	if parsed.ID != id || parsed.BranchCode != "BLR-HUB-01" || *parsed.DistanceKM != 12.34 {
		t.Errorf("unmarshaled branch mismatch: %+v", parsed)
	}
}

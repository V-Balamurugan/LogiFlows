package transfers_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/transfers"
)

func TestTransfers_Validation(t *testing.T) {
	validReq := transfers.CreateBranchTransferRequest{
		SourceBranchID:      "11111111-1111-1111-1111-111111111111",
		DestinationBranchID: "22222222-2222-2222-2222-222222222222",
		ParcelIDs:           []string{"33333333-3333-3333-3333-333333333333"},
	}

	if err := transfers.ValidateCreateRequest(validReq); err != nil {
		t.Fatalf("expected valid create request to pass, got: %v", err)
	}

	// Same branch
	invalidReq := validReq
	invalidReq.DestinationBranchID = invalidReq.SourceBranchID
	if err := transfers.ValidateCreateRequest(invalidReq); err != transfers.ErrSameBranchTransfer {
		t.Errorf("expected ErrSameBranchTransfer, got: %v", err)
	}

	// No parcels
	invalidReq = validReq
	invalidReq.ParcelIDs = nil
	if err := transfers.ValidateCreateRequest(invalidReq); err != transfers.ErrNoParcelsSpecified {
		t.Errorf("expected ErrNoParcelsSpecified, got: %v", err)
	}
}

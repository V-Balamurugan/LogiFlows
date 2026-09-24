package deliveries_test

import (
	"testing"

	"github.com/logiflows/logiflows/backend/internal/deliveries"
)

func TestDelivery_ValidationRules(t *testing.T) {
	validReq := deliveries.CreateDeliveryTaskRequest{
		ParcelID:         "11111111-1111-1111-1111-111111111111",
		AssignedDriverID: "22222222-2222-2222-2222-222222222222",
	}

	if err := deliveries.ValidateCreateRequest(validReq); err != nil {
		t.Fatalf("expected valid create request to pass, got: %v", err)
	}

	// Missing parcel ID
	invalidReq := validReq
	invalidReq.ParcelID = ""
	if err := deliveries.ValidateCreateRequest(invalidReq); err == nil {
		t.Errorf("expected error for missing parcel_id")
	}

	// Missing driver ID
	invalidReq = validReq
	invalidReq.AssignedDriverID = ""
	if err := deliveries.ValidateCreateRequest(invalidReq); err == nil {
		t.Errorf("expected error for missing assigned_driver_id")
	}

	// Invalid priority
	invalidPriority := "SUPER_DUPER_URGENT"
	invalidReq = validReq
	invalidReq.Priority = &invalidPriority
	if err := deliveries.ValidateCreateRequest(invalidReq); err == nil {
		t.Errorf("expected error for invalid priority")
	}
}

func TestDelivery_ProofValidation(t *testing.T) {
	validProof := deliveries.SubmitDeliveryProofRequest{
		ProofType:     deliveries.ProofTypeRecipientSignature,
		RecipientName: "John Doe",
	}

	if err := deliveries.ValidateProofRequest(validProof); err != nil {
		t.Fatalf("expected valid proof request to pass, got: %v", err)
	}

	// Missing recipient name
	invalidProof := validProof
	invalidProof.RecipientName = ""
	if err := deliveries.ValidateProofRequest(invalidProof); err != deliveries.ErrMissingRecipientName {
		t.Errorf("expected ErrMissingRecipientName, got: %v", err)
	}

	// Invalid proof type
	invalidProof = validProof
	invalidProof.ProofType = "TELEPATHY"
	if err := deliveries.ValidateProofRequest(invalidProof); err != deliveries.ErrInvalidProofType {
		t.Errorf("expected ErrInvalidProofType, got: %v", err)
	}
}

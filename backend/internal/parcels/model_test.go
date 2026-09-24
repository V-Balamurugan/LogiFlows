package parcels_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/parcels"
)

func TestParcel_ValidationRules(t *testing.T) {
	originID := uuid.New().String()
	destID := uuid.New().String()

	validReq := parcels.CreateParcelRequest{
		SenderName:          "Alice Sender",
		SenderPhone:         "+91 9876543210",
		SenderAddress:       "123 Sender St, Chennai",
		ReceiverName:        "Bob Receiver",
		ReceiverPhone:       "+91 9123456789",
		ReceiverAddress:     "456 Receiver Ave, Bengaluru",
		OriginBranchID:      originID,
		DestinationBranchID: destID,
		WeightKG:            2.5,
		DimensionsCM:        "30x20x15",
		ServiceType:         parcels.ServiceTypeStandard,
	}

	if err := parcels.ValidateCreateRequest(validReq); err != nil {
		t.Fatalf("expected valid request to pass, got: %v", err)
	}

	// Test missing sender
	invalidReq := validReq
	invalidReq.SenderName = ""
	if err := parcels.ValidateCreateRequest(invalidReq); err != parcels.ErrInvalidSenderDetails {
		t.Errorf("expected ErrInvalidSenderDetails, got: %v", err)
	}

	// Test missing receiver
	invalidReq = validReq
	invalidReq.ReceiverName = ""
	if err := parcels.ValidateCreateRequest(invalidReq); err != parcels.ErrInvalidReceiverDetails {
		t.Errorf("expected ErrInvalidReceiverDetails, got: %v", err)
	}

	// Test zero or negative weight
	invalidReq = validReq
	invalidReq.WeightKG = 0
	if err := parcels.ValidateCreateRequest(invalidReq); err != parcels.ErrInvalidWeight {
		t.Errorf("expected ErrInvalidWeight, got: %v", err)
	}

	// Test invalid service type
	invalidReq = validReq
	invalidReq.ServiceType = "DRONE_SUPER_FAST"
	if err := parcels.ValidateCreateRequest(invalidReq); err != parcels.ErrInvalidServiceType {
		t.Errorf("expected ErrInvalidServiceType, got: %v", err)
	}

	// Test same origin and destination
	invalidReq = validReq
	invalidReq.DestinationBranchID = originID
	if err := parcels.ValidateCreateRequest(invalidReq); err != parcels.ErrSameOriginAndDestination {
		t.Errorf("expected ErrSameOriginAndDestination, got: %v", err)
	}
}

func TestParcel_StateMachineTransitions(t *testing.T) {
	// Valid direct transitions
	if !parcels.IsValidTransition(parcels.StatusCreated, parcels.StatusBooked) {
		t.Errorf("expected CREATED -> BOOKED to be valid")
	}
	if !parcels.IsValidTransition(parcels.StatusBooked, parcels.StatusReceivedAtOriginBranch) {
		t.Errorf("expected BOOKED -> RECEIVED_AT_ORIGIN_BRANCH to be valid")
	}
	if !parcels.IsValidTransition(parcels.StatusReceivedAtOriginBranch, parcels.StatusOutForDelivery) {
		t.Errorf("expected RECEIVED_AT_ORIGIN_BRANCH -> OUT_FOR_DELIVERY to be valid")
	}
	if !parcels.IsValidTransition(parcels.StatusOutForDelivery, parcels.StatusDelivered) {
		t.Errorf("expected OUT_FOR_DELIVERY -> DELIVERED to be valid")
	}

	// Idempotent same-state
	if !parcels.IsValidTransition(parcels.StatusOutForDelivery, parcels.StatusOutForDelivery) {
		t.Errorf("expected same-state transition to be valid")
	}

	// Invalid illegal transitions
	if parcels.IsValidTransition(parcels.StatusCreated, parcels.StatusDelivered) {
		t.Errorf("expected CREATED -> DELIVERED to be invalid")
	}
	if parcels.IsValidTransition(parcels.StatusDelivered, parcels.StatusInTransit) {
		t.Errorf("expected DELIVERED -> IN_TRANSIT to be invalid (DELIVERED is terminal)")
	}
	if parcels.IsValidTransition(parcels.StatusCancelled, parcels.StatusOutForDelivery) {
		t.Errorf("expected CANCELLED -> OUT_FOR_DELIVERY to be invalid (CANCELLED is terminal)")
	}
}

func TestParcel_QRPayloadGenerationAndVerification(t *testing.T) {
	tenantID := uuid.New()
	parcelID := uuid.New()
	tracking := "PKG-20260924-0001"

	qrPayload := parcels.GenerateQRPayload(tracking, tenantID, parcelID)
	if qrPayload == "" {
		t.Fatalf("expected non-empty QR payload")
	}

	// Verify valid payload
	parsedTracking, parsedTenant, parsedParcel, err := parcels.VerifyQRPayload(qrPayload)
	if err != nil {
		t.Fatalf("expected valid QR payload to verify cleanly: %v", err)
	}
	if parsedTracking != tracking {
		t.Errorf("expected tracking %s, got %s", tracking, parsedTracking)
	}
	if parsedTenant != tenantID {
		t.Errorf("expected tenant %s, got %s", tenantID, parsedTenant)
	}
	if parsedParcel != parcelID {
		t.Errorf("expected parcel %s, got %s", parcelID, parsedParcel)
	}

	// Verify corrupted/tampered payload
	corrupted := qrPayload[:len(qrPayload)-5] + "aaaaa"
	_, _, _, err = parcels.VerifyQRPayload(corrupted)
	if err == nil {
		t.Errorf("expected error on tampered QR payload, got nil")
	}
}

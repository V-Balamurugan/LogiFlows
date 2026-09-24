package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/logiflows/logiflows/backend/internal/customers"
	"github.com/logiflows/logiflows/backend/internal/parcels"
)

// TC-P4-CUST-001: Customer Management CRUD, Validation & Multi-Tenant Isolation
func TestCustomer_CRUD_And_ParcelBooking_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping customer integration test in short mode")
	}

	router, _ := setupTestRouter(t)
	tokenA, _, tenantIDA := registerTestCompany(t, router, "Apex Logistics", "apex_logistics")
	tokenB, _, tenantIDB := registerTestCompany(t, router, "Beacon Couriers", "beacon_couriers")

	origBranchIDA := createTestBranch(t, router, tokenA, tenantIDA, "Apex Chennai Hub", "Chennai", 13.0827, 80.2707)
	destBranchIDA := createTestBranch(t, router, tokenA, tenantIDA, "Apex Bengaluru Hub", "Bengaluru", 12.9716, 77.5946)

	// 1. Create Sender Customer under Tenant A
	custReqA := customers.CreateCustomerRequest{
		Name:         "Tata Enterprise Hub",
		CompanyName:  ptr("Tata Sons Pvt Ltd"),
		CustomerType: customers.TypeEnterprise,
		Email:        ptr("logistics@tata.com"),
		Phone:        "+91 9840123456",
		AddressLine1: "Bombay House, Homi Mody Street",
		City:         "Mumbai",
		State:        "Maharashtra",
		PostalCode:   "400001",
		Country:      "India",
		ContractTier: customers.TierVIP,
	}
	b, _ := json.Marshal(custReqA)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/customers", tenantIDA), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for customer, got %d: %s", w.Code, w.Body.String())
	}

	var senderResp struct {
		Data customers.Customer `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &senderResp)
	senderCustomer := senderResp.Data

	if senderCustomer.CustomerCode == "" {
		t.Errorf("expected auto-generated customer code, got empty")
	}
	if senderCustomer.ContractTier != customers.TierVIP {
		t.Errorf("expected VIP tier, got %s", senderCustomer.ContractTier)
	}

	// 2. Create Receiver Customer under Tenant A
	custReqReceiver := customers.CreateCustomerRequest{
		Name:         "Infosys Technologies",
		CompanyName:  ptr("Infosys Ltd"),
		CustomerType: customers.TypeBusiness,
		Email:        ptr("procurement@infosys.com"),
		Phone:        "+91 9988776655",
		AddressLine1: "Electronics City, Hosur Road",
		City:         "Bengaluru",
		State:        "Karnataka",
		PostalCode:   "560100",
		Country:      "India",
		ContractTier: customers.TierGold,
	}
	b, _ = json.Marshal(custReqReceiver)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/customers", tenantIDA), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for receiver customer, got %d: %s", w.Code, w.Body.String())
	}

	var receiverResp struct {
		Data customers.Customer `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &receiverResp)
	receiverCustomer := receiverResp.Data

	// 3. Negative Validation Test: Missing required fields
	invalidCustReq := customers.CreateCustomerRequest{
		Name:  "",
		Phone: "",
	}
	b, _ = json.Marshal(invalidCustReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/customers", tenantIDA), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid customer, got %d", w.Code)
	}

	// 4. List Customers with Search Query
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/customers?search=Tata", tenantIDA), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for customer list, got %d", w.Code)
	}

	var listResp struct {
		Data customers.CustomerListResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &listResp)
	if listResp.Data.Total < 1 {
		t.Errorf("expected at least 1 customer matching 'Tata', got %d", listResp.Data.Total)
	}

	// 5. Update Customer details
	updatedCity := "Navi Mumbai"
	updateReq := customers.UpdateCustomerRequest{
		City: &updatedCity,
	}
	b, _ = json.Marshal(updateReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tenants/%s/customers/%s", tenantIDA, senderCustomer.ID), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for customer update, got %d", w.Code)
	}

	// 6. Parcel Booking Linked to Customer with Auto-Fill & Pricing
	senderIDStr := senderCustomer.ID.String()
	receiverIDStr := receiverCustomer.ID.String()
	bookingReq := parcels.CreateParcelRequest{
		SenderCustomerID:    &senderIDStr,
		ReceiverCustomerID:  &receiverIDStr,
		OriginBranchID:      origBranchIDA.String(),
		DestinationBranchID: destBranchIDA.String(),
		WeightKG:            5.0,
		DimensionsCM:        "40x30x20",
		ServiceType:         parcels.ServiceTypeExpress,
	}
	b, _ = json.Marshal(bookingReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/tenants/%s/parcels", tenantIDA), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for booked parcel, got %d: %s", w.Code, w.Body.String())
	}

	var parcelResp struct {
		Data parcels.Parcel `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &parcelResp)
	bookedParcel := parcelResp.Data

	// Verify auto-fill from sender and receiver customer profiles
	if bookedParcel.SenderName != "Tata Enterprise Hub" {
		t.Errorf("expected auto-filled sender name 'Tata Enterprise Hub', got '%s'", bookedParcel.SenderName)
	}
	if bookedParcel.ReceiverName != "Infosys Technologies" {
		t.Errorf("expected auto-filled receiver name 'Infosys Technologies', got '%s'", bookedParcel.ReceiverName)
	}
	if bookedParcel.SenderCustomerID == nil || *bookedParcel.SenderCustomerID != senderCustomer.ID {
		t.Errorf("expected linked sender_customer_id %s, got %v", senderCustomer.ID, bookedParcel.SenderCustomerID)
	}
	if bookedParcel.ReceiverCustomerID == nil || *bookedParcel.ReceiverCustomerID != receiverCustomer.ID {
		t.Errorf("expected linked receiver_customer_id %s, got %v", receiverCustomer.ID, bookedParcel.ReceiverCustomerID)
	}

	// Verify dynamic pricing: Express base (120) + 5kg * 20 (100) = 220.0
	expectedPrice := 220.0
	if bookedParcel.Price != expectedPrice {
		t.Errorf("expected price %.2f, got %.2f", expectedPrice, bookedParcel.Price)
	}

	// 7. Get Customer Associated Parcels
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/customers/%s/parcels", tenantIDA, senderCustomer.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for customer parcels, got %d: %s", w.Code, w.Body.String())
	}

	var custParcelsResp struct {
		Data parcels.ParcelListResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &custParcelsResp)
	if custParcelsResp.Data.Total < 1 {
		t.Errorf("expected at least 1 parcel for customer %s, got %d", senderCustomer.ID, custParcelsResp.Data.Total)
	}

	// 8. Multi-Tenant Isolation Verification:
	// Tenant B attempting to read Tenant A's customer must be rejected with 403 or 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%s/customers/%s", tenantIDB, senderCustomer.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusForbidden {
		t.Errorf("expected 404 or 403 cross-tenant rejection, got %d: %s", w.Code, w.Body.String())
	}
}

func ptr(s string) *string {
	return &s
}

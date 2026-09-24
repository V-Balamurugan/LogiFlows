package parcels

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/logiflows/logiflows/backend/internal/branches"
	"github.com/logiflows/logiflows/backend/internal/customers"
)

type Service interface {
	CreateParcel(ctx context.Context, tenantID, actorID uuid.UUID, req CreateParcelRequest) (*Parcel, error)
	GetParcel(ctx context.Context, tenantID, parcelID uuid.UUID) (*Parcel, error)
	GetParcelByTrackingNumber(ctx context.Context, trackingNumber string) (*Parcel, error)
	ListParcels(ctx context.Context, tenantID uuid.UUID, filter ParcelFilter) (*ParcelListResponse, error)
	UpdateParcel(ctx context.Context, tenantID, parcelID uuid.UUID, req UpdateParcelRequest) (*Parcel, error)
	UpdateParcelStatus(ctx context.Context, tenantID, parcelID, actorID uuid.UUID, actorRole string, req UpdateParcelStatusRequest) (*Parcel, error)
	GetParcelTimeline(ctx context.Context, tenantID, parcelID uuid.UUID) ([]ParcelStatusHistory, error)
	GetParcelQRCode(ctx context.Context, tenantID, parcelID uuid.UUID) (string, error)
	ScanParcel(ctx context.Context, tenantID, actorID uuid.UUID, actorRole string, req ScanParcelRequest) (*ScanParcelResponse, error)
	GetPublicTracking(ctx context.Context, trackingNumber string) (*PublicTrackingResponse, error)
}

type parcelService struct {
	repo         Repository
	branchRepo   branches.Repository
	customerRepo customers.Repository
}

func NewService(repo Repository, branchRepo branches.Repository, customerRepo customers.Repository) Service {
	return &parcelService{
		repo:         repo,
		branchRepo:   branchRepo,
		customerRepo: customerRepo,
	}
}

func (s *parcelService) CreateParcel(ctx context.Context, tenantID, actorID uuid.UUID, req CreateParcelRequest) (*Parcel, error) {
	// Auto-fill sender details from customer profile if customer ID is provided
	var senderCustID *uuid.UUID
	if req.SenderCustomerID != nil && strings.TrimSpace(*req.SenderCustomerID) != "" {
		sID, err := uuid.Parse(strings.TrimSpace(*req.SenderCustomerID))
		if err == nil && sID != uuid.Nil {
			senderCustID = &sID
			if s.customerRepo != nil {
				cust, err := s.customerRepo.GetCustomerByID(ctx, tenantID, sID)
				if err == nil && cust != nil {
					if strings.TrimSpace(req.SenderName) == "" {
						req.SenderName = cust.Name
					}
					if strings.TrimSpace(req.SenderPhone) == "" {
						req.SenderPhone = cust.Phone
					}
					if strings.TrimSpace(req.SenderAddress) == "" {
						addr := cust.AddressLine1
						if cust.AddressLine2 != nil && *cust.AddressLine2 != "" {
							addr += ", " + *cust.AddressLine2
						}
						addr += ", " + cust.City + ", " + cust.State + " " + cust.PostalCode
						req.SenderAddress = addr
					}
					if req.SenderEmail == nil && cust.Email != nil {
						req.SenderEmail = cust.Email
					}
				}
			}
		}
	}

	// Auto-fill receiver details from customer profile if customer ID is provided
	var receiverCustID *uuid.UUID
	if req.ReceiverCustomerID != nil && strings.TrimSpace(*req.ReceiverCustomerID) != "" {
		rID, err := uuid.Parse(strings.TrimSpace(*req.ReceiverCustomerID))
		if err == nil && rID != uuid.Nil {
			receiverCustID = &rID
			if s.customerRepo != nil {
				cust, err := s.customerRepo.GetCustomerByID(ctx, tenantID, rID)
				if err == nil && cust != nil {
					if strings.TrimSpace(req.ReceiverName) == "" {
						req.ReceiverName = cust.Name
					}
					if strings.TrimSpace(req.ReceiverPhone) == "" {
						req.ReceiverPhone = cust.Phone
					}
					if strings.TrimSpace(req.ReceiverAddress) == "" {
						addr := cust.AddressLine1
						if cust.AddressLine2 != nil && *cust.AddressLine2 != "" {
							addr += ", " + *cust.AddressLine2
						}
						addr += ", " + cust.City + ", " + cust.State + " " + cust.PostalCode
						req.ReceiverAddress = addr
					}
					if req.ReceiverEmail == nil && cust.Email != nil {
						req.ReceiverEmail = cust.Email
					}
				}
			}
		}
	}

	if strings.TrimSpace(req.DimensionsCM) == "" {
		req.DimensionsCM = "30x20x15"
	}
	if strings.TrimSpace(req.ServiceType) == "" {
		req.ServiceType = ServiceTypeStandard
	}

	if err := ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	origID, err := uuid.Parse(strings.TrimSpace(req.OriginBranchID))
	if err != nil {
		return nil, ErrInvalidBranch
	}
	destID, err := uuid.Parse(strings.TrimSpace(req.DestinationBranchID))
	if err != nil {
		return nil, ErrInvalidBranch
	}

	// Verify origin branch belongs to tenant
	origBranch, err := s.branchRepo.GetByID(ctx, tenantID, origID)
	if err != nil || origBranch.TenantID != tenantID {
		return nil, ErrBranchCrossTenant
	}

	// Verify destination branch belongs to tenant
	destBranch, err := s.branchRepo.GetByID(ctx, tenantID, destID)
	if err != nil || destBranch.TenantID != tenantID {
		return nil, ErrBranchCrossTenant
	}

	// Auto-generate sequential tracking number if not specified
	trackingNum := strings.TrimSpace(req.TrackingNumber)
	if trackingNum == "" {
		generated, err := s.repo.NextTrackingNumber(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tracking number: %w", err)
		}
		trackingNum = generated
	}

	declaredVal := 0.0
	if req.DeclaredValue != nil && *req.DeclaredValue >= 0 {
		declaredVal = *req.DeclaredValue
	}

	var createdBy *uuid.UUID
	if actorID != uuid.Nil {
		createdBy = &actorID
	}

	price := 0.0
	if req.Price != nil && *req.Price >= 0 {
		price = *req.Price
	} else {
		price = CalculateEstimatedPrice(req.ServiceType, req.WeightKG, declaredVal)
	}

	parcel := &Parcel{
		TenantID:            tenantID,
		TrackingNumber:      trackingNum,
		SenderCustomerID:    senderCustID,
		SenderName:          strings.TrimSpace(req.SenderName),
		SenderPhone:         strings.TrimSpace(req.SenderPhone),
		SenderEmail:         req.SenderEmail,
		SenderAddress:       strings.TrimSpace(req.SenderAddress),
		ReceiverCustomerID:  receiverCustID,
		ReceiverName:        strings.TrimSpace(req.ReceiverName),
		ReceiverPhone:       strings.TrimSpace(req.ReceiverPhone),
		ReceiverEmail:       req.ReceiverEmail,
		ReceiverAddress:     strings.TrimSpace(req.ReceiverAddress),
		OriginBranchID:      origID,
		DestinationBranchID: destID,
		CurrentBranchID:     &origID,
		WeightKG:            req.WeightKG,
		DimensionsCM:        strings.TrimSpace(req.DimensionsCM),
		ServiceType:         strings.ToUpper(strings.TrimSpace(req.ServiceType)),
		DeclaredValue:       declaredVal,
		Price:               price,
		Status:              StatusCreated,
		SpecialInstructions: req.SpecialInstructions,
		CreatedBy:           createdBy,
	}

	if err := s.repo.CreateParcel(ctx, parcel); err != nil {
		return nil, err
	}

	// Fetch full joined details
	return s.repo.GetParcelByID(ctx, tenantID, parcel.ID)
}

func (s *parcelService) GetParcel(ctx context.Context, tenantID, parcelID uuid.UUID) (*Parcel, error) {
	return s.repo.GetParcelByID(ctx, tenantID, parcelID)
}

func (s *parcelService) GetParcelByTrackingNumber(ctx context.Context, trackingNumber string) (*Parcel, error) {
	return s.repo.GetParcelByTrackingNumber(ctx, trackingNumber)
}

func (s *parcelService) ListParcels(ctx context.Context, tenantID uuid.UUID, filter ParcelFilter) (*ParcelListResponse, error) {
	items, total, err := s.repo.ListParcels(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	return &ParcelListResponse{
		Parcels: items,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	}, nil
}

func (s *parcelService) UpdateParcel(ctx context.Context, tenantID, parcelID uuid.UUID, req UpdateParcelRequest) (*Parcel, error) {
	existing, err := s.repo.GetParcelByID(ctx, tenantID, parcelID)
	if err != nil {
		return nil, err
	}

	if req.SenderName != nil && strings.TrimSpace(*req.SenderName) != "" {
		existing.SenderName = strings.TrimSpace(*req.SenderName)
	}
	if req.SenderPhone != nil && strings.TrimSpace(*req.SenderPhone) != "" {
		existing.SenderPhone = strings.TrimSpace(*req.SenderPhone)
	}
	if req.SenderEmail != nil {
		existing.SenderEmail = req.SenderEmail
	}
	if req.SenderAddress != nil && strings.TrimSpace(*req.SenderAddress) != "" {
		existing.SenderAddress = strings.TrimSpace(*req.SenderAddress)
	}
	if req.ReceiverName != nil && strings.TrimSpace(*req.ReceiverName) != "" {
		existing.ReceiverName = strings.TrimSpace(*req.ReceiverName)
	}
	if req.ReceiverPhone != nil && strings.TrimSpace(*req.ReceiverPhone) != "" {
		existing.ReceiverPhone = strings.TrimSpace(*req.ReceiverPhone)
	}
	if req.ReceiverEmail != nil {
		existing.ReceiverEmail = req.ReceiverEmail
	}
	if req.ReceiverAddress != nil && strings.TrimSpace(*req.ReceiverAddress) != "" {
		existing.ReceiverAddress = strings.TrimSpace(*req.ReceiverAddress)
	}
	if req.WeightKG != nil && *req.WeightKG > 0 {
		existing.WeightKG = *req.WeightKG
	}
	if req.DimensionsCM != nil && strings.TrimSpace(*req.DimensionsCM) != "" {
		existing.DimensionsCM = strings.TrimSpace(*req.DimensionsCM)
	}
	if req.ServiceType != nil && strings.TrimSpace(*req.ServiceType) != "" {
		st := strings.ToUpper(strings.TrimSpace(*req.ServiceType))
		existing.ServiceType = st
	}
	if req.DeclaredValue != nil && *req.DeclaredValue >= 0 {
		existing.DeclaredValue = *req.DeclaredValue
	}
	if req.SpecialInstructions != nil {
		existing.SpecialInstructions = req.SpecialInstructions
	}

	if err := s.repo.UpdateParcel(ctx, existing); err != nil {
		return nil, err
	}

	return s.repo.GetParcelByID(ctx, tenantID, parcelID)
}

func (s *parcelService) UpdateParcelStatus(ctx context.Context, tenantID, parcelID, actorID uuid.UUID, actorRole string, req UpdateParcelStatusRequest) (*Parcel, error) {
	toStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if toStatus == "" {
		return nil, ErrInvalidStatus
	}

	var branchUUID *uuid.UUID
	if req.BranchID != nil && strings.TrimSpace(*req.BranchID) != "" {
		bID, err := uuid.Parse(strings.TrimSpace(*req.BranchID))
		if err != nil {
			return nil, ErrInvalidBranch
		}
		b, err := s.branchRepo.GetByID(ctx, tenantID, bID)
		if err != nil || b.TenantID != tenantID {
			return nil, ErrBranchCrossTenant
		}
		branchUUID = &bID
	}

	notes := ""
	if req.Notes != nil {
		notes = strings.TrimSpace(*req.Notes)
	}

	var actorPtr *uuid.UUID
	if actorID != uuid.Nil {
		actorPtr = &actorID
	}

	if err := s.repo.UpdateParcelStatus(ctx, tenantID, parcelID, toStatus, branchUUID, actorPtr, actorRole, notes); err != nil {
		return nil, err
	}

	return s.repo.GetParcelByID(ctx, tenantID, parcelID)
}

func (s *parcelService) GetParcelTimeline(ctx context.Context, tenantID, parcelID uuid.UUID) ([]ParcelStatusHistory, error) {
	// Verify parcel exists under tenant first
	if _, err := s.repo.GetParcelByID(ctx, tenantID, parcelID); err != nil {
		return nil, err
	}

	return s.repo.GetParcelTimeline(ctx, tenantID, parcelID)
}

func (s *parcelService) GetParcelQRCode(ctx context.Context, tenantID, parcelID uuid.UUID) (string, error) {
	p, err := s.repo.GetParcelByID(ctx, tenantID, parcelID)
	if err != nil {
		return "", err
	}

	if p.QRCodePayload != nil && *p.QRCodePayload != "" {
		return *p.QRCodePayload, nil
	}

	payload := GenerateQRPayload(p.TrackingNumber, tenantID, parcelID)
	return payload, nil
}

func (s *parcelService) ScanParcel(ctx context.Context, tenantID, actorID uuid.UUID, actorRole string, req ScanParcelRequest) (*ScanParcelResponse, error) {
	var tracking string

	if req.QRPayload != nil && strings.TrimSpace(*req.QRPayload) != "" {
		t, qrTenantID, _, err := VerifyQRPayload(strings.TrimSpace(*req.QRPayload))
		if err != nil {
			return nil, ErrInvalidQRPayload
		}
		if qrTenantID != tenantID {
			return nil, ErrParcelNotFound
		}
		tracking = t
	} else if req.TrackingNumber != nil && strings.TrimSpace(*req.TrackingNumber) != "" {
		tracking = strings.TrimSpace(*req.TrackingNumber)
	} else {
		return nil, errors.New("either qr_payload or tracking_number is required")
	}

	p, err := s.repo.GetParcelByTrackingNumber(ctx, tracking)
	if err != nil {
		return nil, err
	}
	if p.TenantID != tenantID {
		return nil, ErrParcelNotFound
	}

	// Determine next recommended operational action
	nextAction := "VERIFIED"
	switch p.Status {
	case StatusCreated, StatusBooked:
		nextAction = "READY_FOR_INTAKE_SCAN"
	case StatusReadyForPickup, StatusPickedUp:
		nextAction = "RECEIVE_AT_ORIGIN_HUB"
	case StatusReceivedAtOriginBranch, StatusReceivedAtTransferBranch:
		nextAction = "ASSIGN_FOR_DELIVERY_OR_TRANSFER"
	case StatusInTransit:
		nextAction = "RECEIVE_AT_DESTINATION_HUB"
	case StatusOutForDelivery:
		nextAction = "CONFIRM_DELIVERY_WITH_POD"
	case StatusDelivered:
		nextAction = "DELIVERED_COMPLETED"
	}

	// Log custody scan event
	var bID *uuid.UUID
	if req.BranchID != nil && strings.TrimSpace(*req.BranchID) != "" {
		if id, err := uuid.Parse(strings.TrimSpace(*req.BranchID)); err == nil {
			bID = &id
		}
	}
	if bID == nil {
		bID = p.CurrentBranchID
	}

	var empID *uuid.UUID
	if actorID != uuid.Nil {
		empID = &actorID
	}

	custodyEvent := &ParcelCustodyEvent{
		TenantID:         tenantID,
		ParcelID:         p.ID,
		EmployeeID:       empID,
		ToBranchID:       bID,
		EventType:        CustodyEventReceive,
		SignatureNote:    req.Notes,
		VerificationCode: &tracking,
	}
	_ = s.repo.RecordCustodyEvent(ctx, custodyEvent)

	return &ScanParcelResponse{
		Parcel:        p,
		VerifiedAt:    time.Now().UTC(),
		CurrentBranch: p.CurrentBranchName,
		NextAction:    nextAction,
		CustodyHolder: p.CurrentBranchName,
	}, nil
}

func (s *parcelService) GetPublicTracking(ctx context.Context, trackingNumber string) (*PublicTrackingResponse, error) {
	cleanTracking := strings.TrimSpace(trackingNumber)
	if cleanTracking == "" {
		return nil, ErrParcelNotFound
	}

	p, err := s.repo.GetParcelByTrackingNumber(ctx, cleanTracking)
	if err != nil {
		return nil, ErrParcelNotFound
	}

	history, err := s.repo.GetPublicTimeline(ctx, cleanTracking)
	if err != nil {
		history = make([]ParcelStatusHistory, 0)
	}

	// Sanitize origin / destination city from address
	originCity := extractCity(p.SenderAddress)
	destCity := extractCity(p.ReceiverAddress)

	milestones := make([]PublicTrackingMilestone, 0, len(history))
	for _, h := range history {
		loc := "In Transit"
		if h.BranchName != nil && *h.BranchName != "" {
			loc = *h.BranchName
		}
		milestones = append(milestones, PublicTrackingMilestone{
			Status:      h.ToStatus,
			Description: formatStatusDescription(h.ToStatus),
			Location:    loc,
			Timestamp:   h.CreatedAt,
		})
	}

	// Calculate estimated delivery if active
	var estDelivery *time.Time
	if p.Status != StatusDelivered && p.Status != StatusCancelled && p.Status != StatusReturned {
		est := p.CreatedAt.Add(48 * time.Hour)
		estDelivery = &est
	}

	return &PublicTrackingResponse{
		TrackingNumber:    p.TrackingNumber,
		Status:            p.Status,
		ServiceType:       p.ServiceType,
		OriginCity:        originCity,
		DestinationCity:   destCity,
		WeightKG:          p.WeightKG,
		CreatedAt:         p.CreatedAt,
		EstimatedDelivery: estDelivery,
		Milestones:        milestones,
	}, nil
}

func extractCity(address string) string {
	parts := strings.Split(address, ",")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return strings.TrimSpace(address)
}

func formatStatusDescription(status string) string {
	switch status {
	case StatusCreated:
		return "Shipment order booked"
	case StatusBooked:
		return "Shipment confirmed & pickup scheduled"
	case StatusReadyForPickup:
		return "Ready for courier pickup"
	case StatusPickedUp:
		return "Package collected from sender"
	case StatusReceivedAtOriginBranch:
		return "Package processed at origin sorting hub"
	case StatusInTransit:
		return "Shipment in transit between hubs"
	case StatusReceivedAtTransferBranch:
		return "Arrived at destination delivery hub"
	case StatusOutForDelivery:
		return "Out for delivery with courier executive"
	case StatusDeliveryAttempted:
		return "Delivery attempted, recipient unavailable"
	case StatusDelivered:
		return "Delivered successfully"
	case StatusDeliveryFailed:
		return "Delivery unsuccessful"
	case StatusReturnInitiated:
		return "Return to origin initiated"
	case StatusReturned:
		return "Package returned to merchant"
	case StatusCancelled:
		return "Order cancelled"
	case StatusOnHold:
		return "Package on hold for verification"
	default:
		return status
	}
}

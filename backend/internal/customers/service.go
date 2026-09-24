package customers

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	CreateCustomer(ctx context.Context, tenantID uuid.UUID, req CreateCustomerRequest) (*Customer, error)
	GetCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*Customer, error)
	GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Customer, error)
	ListCustomers(ctx context.Context, tenantID uuid.UUID, filter CustomerFilter) (*CustomerListResponse, error)
	UpdateCustomer(ctx context.Context, tenantID, customerID uuid.UUID, req UpdateCustomerRequest) (*Customer, error)
	DeleteCustomer(ctx context.Context, tenantID, customerID uuid.UUID) error
}

type customerService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &customerService{repo: repo}
}

func (s *customerService) CreateCustomer(ctx context.Context, tenantID uuid.UUID, req CreateCustomerRequest) (*Customer, error) {
	if err := ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(req.CustomerCode)
	if code == "" {
		generated, err := s.repo.NextCustomerCode(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		code = generated
	}

	cType := strings.ToUpper(strings.TrimSpace(req.CustomerType))
	if cType == "" {
		cType = TypeIndividual
	}

	country := strings.TrimSpace(req.Country)
	if country == "" {
		country = "India"
	}

	tier := strings.ToUpper(strings.TrimSpace(req.ContractTier))
	if tier == "" {
		tier = TierStandard
	}

	creditLimit := 0.0
	if req.CreditLimit != nil && *req.CreditLimit >= 0 {
		creditLimit = *req.CreditLimit
	}

	customer := &Customer{
		TenantID:     tenantID,
		CustomerCode: code,
		Name:         strings.TrimSpace(req.Name),
		CompanyName:  req.CompanyName,
		CustomerType: cType,
		Email:        req.Email,
		Phone:        strings.TrimSpace(req.Phone),
		AddressLine1: strings.TrimSpace(req.AddressLine1),
		AddressLine2: req.AddressLine2,
		City:         strings.TrimSpace(req.City),
		State:        strings.TrimSpace(req.State),
		PostalCode:   strings.TrimSpace(req.PostalCode),
		Country:      country,
		Status:       StatusActive,
		CreditLimit:  creditLimit,
		ContractTier: tier,
		Notes:        req.Notes,
	}

	if err := s.repo.CreateCustomer(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *customerService) GetCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*Customer, error) {
	return s.repo.GetCustomerByID(ctx, tenantID, customerID)
}

func (s *customerService) GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Customer, error) {
	cleanCode := strings.TrimSpace(code)
	if cleanCode == "" {
		return nil, ErrCustomerNotFound
	}
	return s.repo.GetCustomerByCode(ctx, tenantID, cleanCode)
}

func (s *customerService) ListCustomers(ctx context.Context, tenantID uuid.UUID, filter CustomerFilter) (*CustomerListResponse, error) {
	customers, total, err := s.repo.ListCustomers(ctx, tenantID, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	return &CustomerListResponse{
		Customers: customers,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func (s *customerService) UpdateCustomer(ctx context.Context, tenantID, customerID uuid.UUID, req UpdateCustomerRequest) (*Customer, error) {
	existing, err := s.repo.GetCustomerByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.CompanyName != nil {
		existing.CompanyName = req.CompanyName
	}
	if req.CustomerType != nil && strings.TrimSpace(*req.CustomerType) != "" {
		cType := strings.ToUpper(strings.TrimSpace(*req.CustomerType))
		validTypes := map[string]bool{
			TypeIndividual: true,
			TypeBusiness:   true,
			TypeEnterprise: true,
			TypeMerchant:   true,
		}
		if !validTypes[cType] {
			return nil, ErrInvalidCustomerType
		}
		existing.CustomerType = cType
	}
	if req.Email != nil {
		existing.Email = req.Email
	}
	if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		existing.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.AddressLine1 != nil && strings.TrimSpace(*req.AddressLine1) != "" {
		existing.AddressLine1 = strings.TrimSpace(*req.AddressLine1)
	}
	if req.AddressLine2 != nil {
		existing.AddressLine2 = req.AddressLine2
	}
	if req.City != nil && strings.TrimSpace(*req.City) != "" {
		existing.City = strings.TrimSpace(*req.City)
	}
	if req.State != nil && strings.TrimSpace(*req.State) != "" {
		existing.State = strings.TrimSpace(*req.State)
	}
	if req.PostalCode != nil && strings.TrimSpace(*req.PostalCode) != "" {
		existing.PostalCode = strings.TrimSpace(*req.PostalCode)
	}
	if req.Country != nil && strings.TrimSpace(*req.Country) != "" {
		existing.Country = strings.TrimSpace(*req.Country)
	}
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		st := strings.ToUpper(strings.TrimSpace(*req.Status))
		if st != StatusActive && st != StatusInactive && st != StatusSuspended {
			return nil, ErrInvalidCustomerStatus
		}
		existing.Status = st
	}
	if req.CreditLimit != nil && *req.CreditLimit >= 0 {
		existing.CreditLimit = *req.CreditLimit
	}
	if req.ContractTier != nil && strings.TrimSpace(*req.ContractTier) != "" {
		existing.ContractTier = strings.ToUpper(strings.TrimSpace(*req.ContractTier))
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	if err := s.repo.UpdateCustomer(ctx, existing); err != nil {
		return nil, err
	}

	return s.repo.GetCustomerByID(ctx, tenantID, customerID)
}

func (s *customerService) DeleteCustomer(ctx context.Context, tenantID, customerID uuid.UUID) error {
	return s.repo.DeleteCustomer(ctx, tenantID, customerID)
}

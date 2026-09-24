export type CustomerType = 'INDIVIDUAL' | 'BUSINESS' | 'ENTERPRISE' | 'MERCHANT';
export type CustomerStatus = 'ACTIVE' | 'INACTIVE' | 'SUSPENDED';

export interface Customer {
  id: string;
  tenant_id: string;
  customer_code: string;
  customer_type: CustomerType;
  name: string;
  company_name?: string;
  email?: string;
  phone: string;
  tax_id?: string;
  billing_address: string;
  shipping_address?: string;
  status: CustomerStatus;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface CreateCustomerPayload {
  customer_type: CustomerType;
  name: string;
  company_name?: string;
  email?: string;
  phone: string;
  tax_id?: string;
  billing_address: string;
  shipping_address?: string;
  notes?: string;
}

export interface UpdateCustomerPayload {
  customer_type?: CustomerType;
  name?: string;
  company_name?: string;
  email?: string;
  phone?: string;
  tax_id?: string;
  billing_address?: string;
  shipping_address?: string;
  status?: CustomerStatus;
  notes?: string;
}

export interface CustomerFilter {
  search?: string;
  customer_type?: string;
  status?: string;
  page?: number;
  limit?: number;
}

export interface CustomerListResponse {
  customers: Customer[];
  total: number;
  page: number;
  limit: number;
}

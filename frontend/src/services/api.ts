import type { 
  AuthResponse, 
  CurrentUserResponse, 
  MemberDetails, 
  ApiSuccessEnvelope, 
  ApiErrorEnvelope 
} from '../types/auth';
import type {
  Parcel,
  ParcelListResponse,
  CreateParcelPayload,
  UpdateParcelPayload,
  ParcelStatusHistory,
  DeliveryTask,
  CreateDeliveryTaskPayload,
  RecordDeliveryAttemptPayload,
  SubmitDeliveryProofPayload,
  DeliveryAttempt,
  DeliveryProof,
  BranchTransfer,
  CreateBranchTransferPayload,
  PublicTrackingResponse,
  ScanParcelResponse,
} from '../types/parcels';
import type {
  Customer,
  CreateCustomerPayload,
  UpdateCustomerPayload,
  CustomerFilter,
  CustomerListResponse,
} from '../types/customers';

const API_BASE = 'http://localhost:8080/api/v1';
const TOKEN_KEY = 'logiflows_access_token';
const REFRESH_TOKEN_KEY = 'logiflows_refresh_token';

export function getSavedToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setSavedToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getSavedRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setSavedRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_TOKEN_KEY, token);
}

export function clearSavedTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

// Backward compatibility alias
export const clearSavedToken = clearSavedTokens;

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getSavedToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  const contentType = response.headers.get('content-type');
  const isJson = contentType && contentType.includes('application/json');
  const body = isJson ? await response.json() : null;

  if (!response.ok) {
    const errorEnv = body as ApiErrorEnvelope;
    const msg = errorEnv?.error?.message || `HTTP ${response.status} Request failed`;
    const code = errorEnv?.error?.code || 'UNKNOWN_ERROR';
    const err = new Error(msg);
    (err as any).code = code;
    (err as any).status = response.status;
    throw err;
  }

  const envelope = body as ApiSuccessEnvelope<T>;
  return envelope.data;
}

export const api = {
  // Auth
  register: async (payload: { email: string; password: string; full_name: string; phone_number?: string; company_name: string }): Promise<AuthResponse> => {
    const data = await request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    if (data.token) setSavedToken(data.token);
    if (data.refresh_token) setSavedRefreshToken(data.refresh_token);
    return data;
  },

  login: async (payload: { email: string; password: string }): Promise<AuthResponse> => {
    const data = await request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    if (data.token) setSavedToken(data.token);
    if (data.refresh_token) setSavedRefreshToken(data.refresh_token);
    return data;
  },

  refresh: async (): Promise<AuthResponse> => {
    const refreshToken = getSavedRefreshToken();
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }
    const data = await request<AuthResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (data.token) setSavedToken(data.token);
    if (data.refresh_token) setSavedRefreshToken(data.refresh_token);
    return data;
  },

  getMe: () =>
    request<CurrentUserResponse>('/auth/me', {
      method: 'GET',
    }),

  logout: async () => {
    const refreshToken = getSavedRefreshToken();
    try {
      await request('/auth/logout', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken || '' }),
      });
    } finally {
      clearSavedTokens();
    }
  },

  // Tenants
  listMembers: (tenantId: string) =>
    request<MemberDetails[]>(`/tenants/${tenantId}/members`, {
      method: 'GET',
    }),

  addMember: (tenantId: string, email: string, role: string) =>
    request<MemberDetails>(`/tenants/${tenantId}/members`, {
      method: 'POST',
      body: JSON.stringify({ email, role }),
    }),

  getTenant: (tenantId: string) =>
    request<{ id: string; name: string; slug: string; status: string; contact_email: string; created_at?: string }>(`/tenants/${tenantId}`, {
      method: 'GET',
    }),

  getCurrentTenant: () =>
    request<{ id: string; name: string; slug: string; status: string; contact_email: string; created_at?: string }>('/tenants/current', {
      method: 'GET',
    }),

  updateTenant: (tenantId: string, payload: { name?: string; contact_email?: string }) =>
    request<{ id: string; name: string; slug: string; status: string; contact_email: string }>(`/tenants/${tenantId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),

  // Branches
  listBranches: (tenantId: string, params?: { search?: string; status?: string; page?: number; limit?: number; near_lat?: number; near_lng?: number; radius_km?: number }) => {
    const query = new URLSearchParams();
    if (params?.search) query.set('search', params.search);
    if (params?.status) query.set('status', params.status);
    if (params?.page) query.set('page', String(params.page));
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.near_lat !== undefined) query.set('near_lat', String(params.near_lat));
    if (params?.near_lng !== undefined) query.set('near_lng', String(params.near_lng));
    if (params?.radius_km !== undefined) query.set('radius_km', String(params.radius_km));

    const qs = query.toString();
    const endpoint = `/tenants/${tenantId}/branches${qs ? `?${qs}` : ''}`;
    return request<{ branches: import('../types/resources').Branch[]; total: number; page: number; limit: number }>(endpoint, {
      method: 'GET',
    });
  },

  getBranch: (tenantId: string, branchId: string) =>
    request<import('../types/resources').Branch>(`/tenants/${tenantId}/branches/${branchId}`, {
      method: 'GET',
    }),

  createBranch: (tenantId: string, payload: import('../types/resources').CreateBranchPayload) =>
    request<import('../types/resources').Branch>(`/tenants/${tenantId}/branches`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateBranch: (tenantId: string, branchId: string, payload: Partial<import('../types/resources').CreateBranchPayload> & { status?: string }) =>
    request<import('../types/resources').Branch>(`/tenants/${tenantId}/branches/${branchId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),

  updateBranchStatus: (tenantId: string, branchId: string, status: string) =>
    request<import('../types/resources').Branch>(`/tenants/${tenantId}/branches/${branchId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),

  deleteBranch: (tenantId: string, branchId: string) =>
    request<{ message: string }>(`/tenants/${tenantId}/branches/${branchId}`, {
      method: 'DELETE',
    }),

  getBranchEmployees: (tenantId: string, branchId: string) =>
    request<{ employees: import('../types/resources').BranchEmployeeSummary[]; total: number }>(`/tenants/${tenantId}/branches/${branchId}/employees`, {
      method: 'GET',
    }),

  getBranchVehicles: (tenantId: string, branchId: string) =>
    request<{ vehicles: import('../types/resources').BranchVehicleSummary[]; total: number }>(`/tenants/${tenantId}/branches/${branchId}/vehicles`, {
      method: 'GET',
    }),

  // Employees
  listEmployees: (tenantId: string, params?: { search?: string; operational_role?: string; branch_id?: string; status?: string; limit?: number; offset?: number }) => {
    const query = new URLSearchParams();
    if (params?.search) query.set('search', params.search);
    if (params?.operational_role) query.set('operational_role', params.operational_role);
    if (params?.branch_id) query.set('branch_id', params.branch_id);
    if (params?.status) query.set('status', params.status);
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset !== undefined) query.set('offset', String(params.offset));

    const qs = query.toString();
    const endpoint = `/tenants/${tenantId}/employees${qs ? `?${qs}` : ''}`;
    return request<import('../types/resources').EmployeeListResponse>(endpoint, {
      method: 'GET',
    });
  },

  getEmployee: (tenantId: string, employeeId: string) =>
    request<import('../types/resources').Employee>(`/tenants/${tenantId}/employees/${employeeId}`, {
      method: 'GET',
    }),

  createEmployee: (tenantId: string, payload: import('../types/resources').CreateEmployeePayload) =>
    request<import('../types/resources').Employee>(`/tenants/${tenantId}/employees`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  createEmployeeWithAccount: (tenantId: string, payload: import('../types/resources').CreateEmployeeWithAccountPayload) =>
    request<import('../types/resources').Employee>(`/tenants/${tenantId}/employees/with-account`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  getEmployeeAccountStatus: (tenantId: string, employeeId: string) =>
    request<import('../types/resources').EmployeeAccountStatus>(`/tenants/${tenantId}/employees/${employeeId}/account-status`, {
      method: 'GET',
    }),

  getMyProfile: (tenantId: string) =>
    request<import('../types/resources').EmployeeMeResponse>(`/tenants/${tenantId}/employees/me`, {
      method: 'GET',
    }),

  updateEmployee: (tenantId: string, employeeId: string, payload: Partial<import('../types/resources').CreateEmployeePayload> & { status?: string }) =>
    request<import('../types/resources').Employee>(`/tenants/${tenantId}/employees/${employeeId}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  deleteEmployee: (tenantId: string, employeeId: string) =>
    request<{ message: string }>(`/tenants/${tenantId}/employees/${employeeId}`, {
      method: 'DELETE',
    }),

  updateEmployeeStatus: (tenantId: string, employeeId: string, payload: import('../types/resources').UpdateEmployeeStatusPayload) =>
    request<import('../types/resources').Employee>(`/tenants/${tenantId}/employees/${employeeId}/status`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),

  listAvailableDrivers: (tenantId: string, branchId?: string) => {
    const qs = branchId ? `?branch_id=${encodeURIComponent(branchId)}` : '';
    return request<{ drivers: import('../types/resources').Employee[]; total: number }>(`/tenants/${tenantId}/employees/available-drivers${qs}`, {
      method: 'GET',
    });
  },

  // Vehicles & Fleet Operations
  listVehicles: (tenantId: string, params?: { search?: string; vehicle_type?: string; branch_id?: string; status?: string; limit?: number; offset?: number }) => {
    const query = new URLSearchParams();
    if (params?.search) query.set('search', params.search);
    if (params?.vehicle_type) query.set('vehicle_type', params.vehicle_type);
    if (params?.branch_id) query.set('branch_id', params.branch_id);
    if (params?.status) query.set('status', params.status);
    if (params?.limit) query.set('limit', String(params.limit));
    if (params?.offset !== undefined) query.set('offset', String(params.offset));

    const qs = query.toString();
    const endpoint = `/tenants/${tenantId}/vehicles${qs ? `?${qs}` : ''}`;
    return request<import('../types/resources').VehicleListResponse>(endpoint, {
      method: 'GET',
    });
  },

  getVehicle: (tenantId: string, vehicleId: string) =>
    request<import('../types/resources').Vehicle>(`/tenants/${tenantId}/vehicles/${vehicleId}`, {
      method: 'GET',
    }),

  createVehicle: (tenantId: string, payload: import('../types/resources').CreateVehiclePayload) =>
    request<import('../types/resources').Vehicle>(`/tenants/${tenantId}/vehicles`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateVehicle: (tenantId: string, vehicleId: string, payload: Partial<import('../types/resources').CreateVehiclePayload> & { status?: string }) =>
    request<import('../types/resources').Vehicle>(`/tenants/${tenantId}/vehicles/${vehicleId}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  updateVehicleStatus: (tenantId: string, vehicleId: string, payload: import('../types/resources').UpdateVehicleStatusPayload) =>
    request<import('../types/resources').Vehicle>(`/tenants/${tenantId}/vehicles/${vehicleId}/status`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),

  deleteVehicle: (tenantId: string, vehicleId: string) =>
    request<{ message: string }>(`/tenants/${tenantId}/vehicles/${vehicleId}`, {
      method: 'DELETE',
    }),

  assignVehicle: (tenantId: string, vehicleId: string, driverId: string, notes?: string) =>
    request<{ id: string; status: string }>(`/tenants/${tenantId}/vehicles/${vehicleId}/assign`, {
      method: 'POST',
      body: JSON.stringify({ driver_id: driverId, notes }),
    }),

  unassignVehicle: (tenantId: string, vehicleId: string) =>
    request<{ message: string }>(`/tenants/${tenantId}/vehicles/${vehicleId}/unassign`, {
      method: 'POST',
    }),

  listAssignments: (tenantId: string, vehicleId?: string) => {
    const qs = vehicleId ? `?vehicle_id=${encodeURIComponent(vehicleId)}` : '';
    return request<{ assignments: import('../types/resources').DriverVehicleAssignment[]; total: number }>(`/tenants/${tenantId}/assignments${qs}`, {
      method: 'GET',
    });
  },

  // --- Phase 4: Parcels ---
  listParcels: (tenantId: string, filter: { status?: string; customer_id?: string; origin_branch_id?: string; destination_branch_id?: string; search?: string; page?: number; limit?: number } = {}) => {
    const params = new URLSearchParams();
    if (filter.status) params.set('status', filter.status);
    if (filter.customer_id) params.set('customer_id', filter.customer_id);
    if (filter.origin_branch_id) params.set('origin_branch_id', filter.origin_branch_id);
    if (filter.destination_branch_id) params.set('destination_branch_id', filter.destination_branch_id);
    if (filter.search) params.set('search', filter.search);
    if (filter.page) params.set('page', String(filter.page));
    if (filter.limit) params.set('limit', String(filter.limit));
    const qs = params.toString() ? `?${params.toString()}` : '';
    return request<ParcelListResponse>(`/tenants/${tenantId}/parcels${qs}`, { method: 'GET' });
  },

  getParcel: (tenantId: string, parcelId: string) =>
    request<Parcel>(`/tenants/${tenantId}/parcels/${parcelId}`, { method: 'GET' }),

  createParcel: (tenantId: string, payload: CreateParcelPayload) =>
    request<Parcel>(`/tenants/${tenantId}/parcels`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateParcel: (tenantId: string, parcelId: string, payload: UpdateParcelPayload) =>
    request<Parcel>(`/tenants/${tenantId}/parcels/${parcelId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),

  updateParcelStatus: (tenantId: string, parcelId: string, status: string, notes?: string, branchId?: string) =>
    request<Parcel>(`/tenants/${tenantId}/parcels/${parcelId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, notes, branch_id: branchId }),
    }),

  getParcelTimeline: (tenantId: string, parcelId: string) =>
    request<ParcelStatusHistory[]>(`/tenants/${tenantId}/parcels/${parcelId}/timeline`, { method: 'GET' }),

  getParcelQRCode: (tenantId: string, parcelId: string) =>
    request<{ parcel_id: string; qr_payload: string }>(`/tenants/${tenantId}/parcels/${parcelId}/qr`, { method: 'GET' }),

  scanParcel: (tenantId: string, payload: { qr_payload?: string; tracking_number?: string; branch_id?: string; notes?: string }) =>
    request<ScanParcelResponse>(`/tenants/${tenantId}/parcels/scan`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  // --- Phase 4: Deliveries ---
  listDeliveryTasks: (tenantId: string, filter: { status?: string; driver_id?: string; priority?: string; page?: number; limit?: number } = {}) => {
    const params = new URLSearchParams();
    if (filter.status) params.set('status', filter.status);
    if (filter.driver_id) params.set('driver_id', filter.driver_id);
    if (filter.priority) params.set('priority', filter.priority);
    if (filter.page) params.set('page', String(filter.page));
    if (filter.limit) params.set('limit', String(filter.limit));
    const qs = params.toString() ? `?${params.toString()}` : '';
    return request<{ tasks: DeliveryTask[]; total: number }>(`/tenants/${tenantId}/deliveries${qs}`, { method: 'GET' });
  },

  getMyDeliveryTasks: (tenantId: string) =>
    request<{ tasks: DeliveryTask[]; total: number }>(`/tenants/${tenantId}/deliveries/my-tasks`, { method: 'GET' }),

  getDeliveryTask: (tenantId: string, deliveryId: string) =>
    request<DeliveryTask>(`/tenants/${tenantId}/deliveries/${deliveryId}`, { method: 'GET' }),

  createDeliveryTask: (tenantId: string, payload: CreateDeliveryTaskPayload) =>
    request<DeliveryTask>(`/tenants/${tenantId}/deliveries`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateDeliveryTaskStatus: (tenantId: string, deliveryId: string, status: string, notes?: string, failureReason?: string) =>
    request<DeliveryTask>(`/tenants/${tenantId}/deliveries/${deliveryId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, notes, failure_reason: failureReason }),
    }),

  recordDeliveryAttempt: (tenantId: string, deliveryId: string, payload: RecordDeliveryAttemptPayload) =>
    request<DeliveryAttempt>(`/tenants/${tenantId}/deliveries/${deliveryId}/attempt`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  getDeliveryAttempts: (tenantId: string, deliveryId: string) =>
    request<DeliveryAttempt[]>(`/tenants/${tenantId}/deliveries/${deliveryId}/attempts`, { method: 'GET' }),

  submitDeliveryProof: (tenantId: string, deliveryId: string, payload: SubmitDeliveryProofPayload) =>
    request<DeliveryProof>(`/tenants/${tenantId}/deliveries/${deliveryId}/proof`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  // --- Phase 4: Branch Transfers ---
  listBranchTransfers: (tenantId: string, filter: { status?: string; origin_branch_id?: string; destination_branch_id?: string; page?: number; limit?: number } = {}) => {
    const params = new URLSearchParams();
    if (filter.status) params.set('status', filter.status);
    if (filter.origin_branch_id) params.set('origin_branch_id', filter.origin_branch_id);
    if (filter.destination_branch_id) params.set('destination_branch_id', filter.destination_branch_id);
    if (filter.page) params.set('page', String(filter.page));
    if (filter.limit) params.set('limit', String(filter.limit));
    const qs = params.toString() ? `?${params.toString()}` : '';
    return request<{ transfers: BranchTransfer[]; total: number }>(`/tenants/${tenantId}/transfers${qs}`, { method: 'GET' });
  },

  getBranchTransfer: (tenantId: string, transferId: string) =>
    request<BranchTransfer>(`/tenants/${tenantId}/transfers/${transferId}`, { method: 'GET' }),

  createBranchTransfer: (tenantId: string, payload: CreateBranchTransferPayload) =>
    request<BranchTransfer>(`/tenants/${tenantId}/transfers`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  dispatchBranchTransfer: (tenantId: string, transferId: string, notes?: string) =>
    request<BranchTransfer>(`/tenants/${tenantId}/transfers/${transferId}/dispatch`, {
      method: 'POST',
      body: JSON.stringify({ notes }),
    }),

  receiveBranchTransfer: (tenantId: string, transferId: string, parcelIds?: string[], notes?: string) =>
    request<BranchTransfer>(`/tenants/${tenantId}/transfers/${transferId}/receive`, {
      method: 'POST',
      body: JSON.stringify({ parcel_ids: parcelIds, notes }),
    }),

  // --- Phase 4: Customer Tracking ---
  getPublicTracking: (trackingNumber: string) =>
    request<PublicTrackingResponse>(`/tracking/${encodeURIComponent(trackingNumber)}`, { method: 'GET' }),

  // --- Phase 4: Customer Management & CRM ---
  listCustomers: (tenantId: string, filter: CustomerFilter = {}) => {
    const params = new URLSearchParams();
    if (filter.search) params.set('search', filter.search);
    if (filter.customer_type) params.set('customer_type', filter.customer_type);
    if (filter.status) params.set('status', filter.status);
    if (filter.page) params.set('page', String(filter.page));
    if (filter.limit) params.set('limit', String(filter.limit));
    const qs = params.toString() ? `?${params.toString()}` : '';
    return request<CustomerListResponse>(`/tenants/${tenantId}/customers${qs}`, { method: 'GET' });
  },

  getCustomer: (tenantId: string, customerId: string) =>
    request<Customer>(`/tenants/${tenantId}/customers/${customerId}`, { method: 'GET' }),

  createCustomer: (tenantId: string, payload: CreateCustomerPayload) =>
    request<Customer>(`/tenants/${tenantId}/customers`, {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateCustomer: (tenantId: string, customerId: string, payload: UpdateCustomerPayload) =>
    request<Customer>(`/tenants/${tenantId}/customers/${customerId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),

  deleteCustomer: (tenantId: string, customerId: string) =>
    request<{ message: string }>(`/tenants/${tenantId}/customers/${customerId}`, {
      method: 'DELETE',
    }),

  getCustomerParcels: (tenantId: string, customerId: string, params?: { page?: number; limit?: number }) => {
    const q = new URLSearchParams();
    if (params?.page) q.set('page', String(params.page));
    if (params?.limit) q.set('limit', String(params.limit));
    const qs = q.toString() ? `?${q.toString()}` : '';
    return request<{ parcels: Parcel[]; total: number; page: number; limit: number }>(`/tenants/${tenantId}/customers/${customerId}/parcels${qs}`, {
      method: 'GET',
    });
  },
};



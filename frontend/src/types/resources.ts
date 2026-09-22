export interface Branch {
  id: string;
  tenant_id: string;
  branch_code: string;
  name: string;
  address: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
  latitude: number;
  longitude: number;
  coverage_radius_km: number;
  status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED';
  is_active: boolean;
  distance_km?: number;
  created_at: string;
  updated_at: string;
}

export interface CreateBranchPayload {
  branch_code: string;
  name: string;
  address: string;
  city: string;
  state?: string;
  postal_code?: string;
  country?: string;
  latitude: number;
  longitude: number;
  coverage_radius_km?: number;
}

export interface BranchListResponse {
  branches: Branch[];
  total: number;
  page: number;
  limit: number;
}

export interface Employee {
  id: string;
  tenant_id: string;
  user_id?: string;
  branch_id?: string;
  branch_name?: string;
  branch_code?: string;
  employee_code: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  designation: string;
  employment_type: 'FULL_TIME' | 'PART_TIME' | 'CONTRACTOR' | 'INTERN';
  operational_role: 'DRIVER' | 'OPERATOR' | 'DISPATCHER' | 'SUPERVISOR' | 'MANAGER';
  license_number?: string;
  status: 'ACTIVE' | 'ON_LEAVE' | 'SUSPENDED' | 'TERMINATED';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateEmployeePayload {
  employee_code: string;
  first_name: string;
  last_name: string;
  designation: string;
  email?: string;
  phone?: string;
  employment_type?: string;
  operational_role?: string;
  license_number?: string;
  user_id?: string;
  branch_id?: string;
}

export interface EmployeeListResponse {
  employees: Employee[];
  total: number;
  limit: number;
  offset: number;
}

export interface Vehicle {
  id: string;
  tenant_id: string;
  assigned_branch_id?: string;
  branch_name?: string;
  branch_code?: string;
  registration_number: string;
  vehicle_type: 'ELECTRIC_VAN' | 'VAN' | 'MOTORCYCLE' | 'TRUCK' | 'THREE_WHEELER';
  make_model?: string;
  year?: number;
  max_weight_kg: number;
  max_volume_cbm: number;
  status: 'AVAILABLE' | 'ASSIGNED' | 'IN_TRANSIT' | 'MAINTENANCE' | 'DECOMMISSIONED';
  is_active: boolean;
  current_driver_name?: string;
  current_driver_id?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateVehiclePayload {
  assigned_branch_id?: string;
  registration_number: string;
  vehicle_type: string;
  make_model?: string;
  year?: number;
  max_weight_kg: number;
  max_volume_cbm: number;
}

export interface VehicleListResponse {
  vehicles: Vehicle[];
  total: number;
  page: number;
  limit: number;
}

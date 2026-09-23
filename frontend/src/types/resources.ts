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
  operational_role: 'DRIVER' | 'OPERATOR' | 'DISPATCHER' | 'SUPERVISOR' | 'MANAGER' | 'BRANCH_MANAGER' | 'WAREHOUSE_OPERATOR' | 'DELIVERY_EXECUTIVE';
  license_number?: string;
  status: 'ACTIVE' | 'ON_LEAVE' | 'SUSPENDED' | 'TERMINATED';
  availability_status: 'AVAILABLE' | 'BUSY' | 'OFF_DUTY' | 'UNAVAILABLE';
  verification_status: 'PENDING' | 'VERIFIED' | 'REJECTED';
  joining_date?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface CreateEmployeePayload {
  employee_code?: string;
  first_name: string;
  last_name: string;
  designation: string;
  email?: string;
  phone?: string;
  employment_type?: string;
  operational_role?: string;
  license_number?: string;
  availability_status?: string;
  verification_status?: string;
  joining_date?: string;
  user_id?: string;
  branch_id?: string;
}

export interface UpdateEmployeeStatusPayload {
  status?: 'ACTIVE' | 'ON_LEAVE' | 'SUSPENDED' | 'TERMINATED';
  availability_status?: 'AVAILABLE' | 'BUSY' | 'OFF_DUTY' | 'UNAVAILABLE';
  verification_status?: 'PENDING' | 'VERIFIED' | 'REJECTED';
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
  availability_status: 'AVAILABLE' | 'BUSY' | 'MAINTENANCE' | 'OUT_OF_SERVICE';
  is_active: boolean;
  current_driver_name?: string;
  current_driver_id?: string;
  current_driver_code?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface CreateVehiclePayload {
  assigned_branch_id?: string;
  registration_number: string;
  vehicle_type: string;
  make_model?: string;
  year?: number;
  max_weight_kg: number;
  max_volume_cbm: number;
  availability_status?: string;
}

export interface UpdateVehicleStatusPayload {
  status?: 'AVAILABLE' | 'ASSIGNED' | 'IN_TRANSIT' | 'MAINTENANCE' | 'DECOMMISSIONED';
  availability_status?: 'AVAILABLE' | 'BUSY' | 'MAINTENANCE' | 'OUT_OF_SERVICE';
}

export interface VehicleListResponse {
  vehicles: Vehicle[];
  total: number;
  page: number;
  limit: number;
}

export interface DriverVehicleAssignment {
  id: string;
  tenant_id: string;
  vehicle_id: string;
  driver_id: string;
  assigned_at: string;
  unassigned_at?: string | null;
  status: 'ACTIVE' | 'COMPLETED' | 'CANCELLED';
  notes?: string;
  driver_name?: string;
  driver_code?: string;
  vehicle_reg?: string;
}

export interface CreateEmployeeWithAccountPayload {
  first_name: string;
  last_name: string;
  email: string;
  designation: string;
  phone?: string;
  employment_type?: string;
  operational_role?: string;
  license_number?: string;
  availability_status?: string;
  verification_status?: string;
  branch_id?: string;
  employee_code?: string;
  password?: string;
  system_role?: string;
  send_invite?: boolean;
}

export interface EmployeeAccountStatus {
  employee_id: string;
  employee_code: string;
  full_name: string;
  operational_role: string;
  status: string;
  has_account: boolean;
  user_id?: string;
  user_email?: string;
  user_is_active?: boolean;
  system_role?: string;
  branch_id?: string;
  branch_name?: string;
}

export interface AssignedVehicleInfo {
  id: string;
  registration_number: string;
  vehicle_type: string;
  make_model?: string;
  status: string;
}

export interface EmployeeMeResponse {
  employee: Employee;
  system_role: string;
  assigned_vehicle?: AssignedVehicleInfo;
}

export interface BranchEmployeeSummary {
  id: string;
  employee_code: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  designation: string;
  operational_role: string;
  status: string;
  availability_status: string;
  is_active: boolean;
}

export interface BranchVehicleSummary {
  id: string;
  registration_number: string;
  vehicle_type: string;
  make_model?: string;
  year?: number;
  max_weight_kg: number;
  max_volume_cbm: number;
  status: string;
  availability_status: string;
  is_active: boolean;
  current_driver_name?: string;
}

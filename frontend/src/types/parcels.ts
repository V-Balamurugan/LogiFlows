export type ParcelStatus =
  | 'CREATED'
  | 'BOOKED'
  | 'READY_FOR_PICKUP'
  | 'PICKED_UP'
  | 'RECEIVED_AT_ORIGIN_BRANCH'
  | 'IN_TRANSIT'
  | 'RECEIVED_AT_TRANSFER_BRANCH'
  | 'OUT_FOR_DELIVERY'
  | 'DELIVERY_ATTEMPTED'
  | 'DELIVERED'
  | 'DELIVERY_FAILED'
  | 'RETURN_INITIATED'
  | 'RETURNED'
  | 'CANCELLED'
  | 'ON_HOLD';

export type ServiceType = 'STANDARD' | 'EXPRESS' | 'OVERNIGHT' | 'SAME_DAY';

export interface Parcel {
  id: string;
  tenant_id: string;
  tracking_number: string;
  sender_name: string;
  sender_phone: string;
  sender_email?: string;
  sender_address: string;
  receiver_name: string;
  receiver_phone: string;
  receiver_email?: string;
  receiver_address: string;
  origin_branch_id: string;
  destination_branch_id: string;
  current_branch_id?: string;
  weight_kg: number;
  dimensions_cm: string;
  service_type: ServiceType;
  declared_value: number;
  status: ParcelStatus;
  special_instructions?: string;
  qr_code_payload?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;

  origin_branch_name?: string;
  origin_branch_code?: string;
  destination_branch_name?: string;
  destination_branch_code?: string;
  current_branch_name?: string;
  current_branch_code?: string;
}

export interface ParcelStatusHistory {
  id: string;
  tenant_id: string;
  parcel_id: string;
  from_status?: ParcelStatus;
  to_status: ParcelStatus;
  branch_id?: string;
  branch_name?: string;
  actor_id?: string;
  actor_name?: string;
  actor_role?: string;
  notes?: string;
  created_at: string;
}

export interface ParcelCustodyEvent {
  id: string;
  tenant_id: string;
  parcel_id: string;
  employee_id?: string;
  employee_name?: string;
  from_branch_id?: string;
  from_branch_name?: string;
  to_branch_id?: string;
  to_branch_name?: string;
  event_type: string;
  signature_note?: string;
  verification_code?: string;
  created_at: string;
}

export interface CreateParcelPayload {
  tracking_number?: string;
  sender_name: string;
  sender_phone: string;
  sender_email?: string;
  sender_address: string;
  receiver_name: string;
  receiver_phone: string;
  receiver_email?: string;
  receiver_address: string;
  origin_branch_id: string;
  destination_branch_id: string;
  weight_kg: number;
  dimensions_cm: string;
  service_type: ServiceType;
  declared_value?: number;
  special_instructions?: string;
}

export interface UpdateParcelPayload {
  sender_name?: string;
  sender_phone?: string;
  sender_email?: string;
  sender_address?: string;
  receiver_name?: string;
  receiver_phone?: string;
  receiver_email?: string;
  receiver_address?: string;
  weight_kg?: number;
  dimensions_cm?: string;
  service_type?: ServiceType;
  declared_value?: number;
  special_instructions?: string;
}

export interface ParcelFilter {
  status?: string;
  origin_branch_id?: string;
  destination_branch_id?: string;
  current_branch_id?: string;
  search?: string;
  page?: number;
  limit?: number;
}

export interface ParcelListResponse {
  parcels: Parcel[];
  total: number;
  page: number;
  limit: number;
}

export type DeliveryTaskStatus = 'ASSIGNED' | 'IN_PROGRESS' | 'COMPLETED' | 'FAILED' | 'CANCELLED';
export type DeliveryPriority = 'LOW' | 'NORMAL' | 'HIGH' | 'URGENT';
export type AttemptOutcome = 'CUSTOMER_UNAVAILABLE' | 'INCORRECT_ADDRESS' | 'REJECTED' | 'SECURITY_RESTRICTED' | 'WEATHER_DELAY' | 'OTHER';
export type ProofType = 'RECIPIENT_SIGNATURE' | 'PHOTO_CONFIRMATION' | 'OTP_VERIFICATION' | 'SECURITY_PASS';

export interface DeliveryTask {
  id: string;
  tenant_id: string;
  parcel_id: string;
  assigned_driver_id: string;
  vehicle_id?: string;
  status: DeliveryTaskStatus;
  priority: DeliveryPriority;
  assigned_at: string;
  started_at?: string;
  completed_at?: string;
  notes?: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;

  tracking_number?: string;
  receiver_name?: string;
  receiver_address?: string;
  receiver_phone?: string;
  driver_name?: string;
  driver_phone?: string;
  vehicle_license_plate?: string;
}

export interface DeliveryAttempt {
  id: string;
  tenant_id: string;
  delivery_task_id: string;
  driver_id: string;
  driver_name?: string;
  attempt_number: number;
  attempt_time: string;
  outcome: AttemptOutcome;
  notes?: string;
  latitude?: number;
  longitude?: number;
}

export interface DeliveryProof {
  id: string;
  tenant_id: string;
  delivery_task_id: string;
  parcel_id: string;
  proof_type: ProofType;
  recipient_name: string;
  recipient_relationship?: string;
  otp_code?: string;
  signature_data?: string;
  photo_url?: string;
  notes?: string;
  latitude?: number;
  longitude?: number;
  verified_at: string;
}

export interface CreateDeliveryTaskPayload {
  parcel_id: string;
  assigned_driver_id: string;
  vehicle_id?: string;
  priority?: DeliveryPriority;
  notes?: string;
}

export interface RecordDeliveryAttemptPayload {
  outcome: AttemptOutcome;
  notes?: string;
  latitude?: number;
  longitude?: number;
}

export interface SubmitDeliveryProofPayload {
  proof_type: ProofType;
  recipient_name: string;
  recipient_relationship?: string;
  otp_code?: string;
  signature_data?: string;
  photo_url?: string;
  notes?: string;
  latitude?: number;
  longitude?: number;
}

export type TransferStatus = 'INITIATED' | 'IN_TRANSIT' | 'RECEIVED' | 'CANCELLED';

export interface BranchTransfer {
  id: string;
  tenant_id: string;
  transfer_number: string;
  origin_branch_id: string;
  destination_branch_id: string;
  responsible_employee_id?: string;
  vehicle_id?: string;
  status: TransferStatus;
  dispatched_at?: string;
  received_at?: string;
  notes?: string;
  created_at: string;
  updated_at: string;

  origin_branch_name?: string;
  destination_branch_name?: string;
  employee_name?: string;
  vehicle_license_plate?: string;
  parcels_count?: number;
  parcels?: Parcel[];
}

export interface CreateBranchTransferPayload {
  origin_branch_id: string;
  destination_branch_id: string;
  responsible_employee_id?: string;
  vehicle_id?: string;
  parcel_ids: string[];
  notes?: string;
}

export interface PublicTrackingResponse {
  tracking_number: string;
  status: ParcelStatus;
  status_display: string;
  service_type: string;
  origin_city: string;
  destination_city: string;
  current_location?: string;
  estimated_delivery?: string;
  created_at: string;
  updated_at: string;
  milestones: {
    status: string;
    description: string;
    location?: string;
    timestamp: string;
  }[];
}

export interface ScanParcelResponse {
  parcel: Parcel;
  verified_at: string;
  current_branch?: string;
  next_action: string;
}

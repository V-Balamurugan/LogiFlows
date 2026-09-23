export interface UserSummary {
  id: string;
  email: string;
  full_name: string;
  phone_number?: string;
  is_active: boolean;
  is_platform_admin: boolean;
  email_verified?: boolean;
}

export interface TenantSummary {
  id: string;
  name: string;
  slug: string;
  role: 'PLATFORM_ADMIN' | 'TENANT_ADMIN' | 'TENANT_OPERATOR' | 'VIEWER' | 'EMPLOYEE';
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  refresh_token?: string;
  refresh_token_expires_at?: string;
  user: UserSummary;
  tenants: TenantSummary[];
}

export interface CurrentUserResponse {
  user: UserSummary;
  tenants: TenantSummary[];
}

export interface MemberDetails {
  id: string;
  tenant_id: string;
  user_id: string;
  email: string;
  full_name: string;
  role: string;
  status: string;
  created_at: string;
}

export interface ApiSuccessEnvelope<T> {
  status: string;
  service: string;
  version: string;
  timestamp: string;
  data: T;
}

export interface ApiErrorEnvelope {
  error: {
    code: string;
    message: string;
    request_id?: string;
    details?: any;
  };
}

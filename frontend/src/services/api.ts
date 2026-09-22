import type { 
  AuthResponse, 
  CurrentUserResponse, 
  MemberDetails, 
  ApiSuccessEnvelope, 
  ApiErrorEnvelope 
} from '../types/auth';

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

  updateTenant: (tenantId: string, payload: { name?: string; contact_email?: string }) =>
    request<{ id: string; name: string; slug: string; status: string; contact_email: string }>(`/tenants/${tenantId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    }),
};

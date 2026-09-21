import React, { createContext, useContext, useState, useEffect } from 'react';
import type { UserSummary, TenantSummary } from '../types/auth';
import { api, getSavedToken, setSavedToken, clearSavedToken } from '../services/api';

interface AuthContextType {
  user: UserSummary | null;
  tenants: TenantSummary[];
  activeTenant: TenantSummary | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  login: (email: string, pass: string) => Promise<void>;
  register: (payload: { email: string; password: string; full_name: string; phone_number?: string; company_name: string }) => Promise<void>;
  logout: () => Promise<void>;
  setActiveTenant: (tenant: TenantSummary) => void;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<UserSummary | null>(null);
  const [tenants, setTenants] = useState<TenantSummary[]>([]);
  const [activeTenant, setActiveTenant] = useState<TenantSummary | null>(null);
  const [token, setToken] = useState<string | null>(getSavedToken());
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Restore authenticated session on mount
  useEffect(() => {
    const initAuth = async () => {
      const saved = getSavedToken();
      if (!saved) {
        setIsLoading(false);
        return;
      }

      try {
        const data = await api.getMe();
        setUser(data.user);
        setTenants(data.tenants);
        if (data.tenants.length > 0) {
          setActiveTenant(data.tenants[0]);
        }
      } catch (err) {
        console.warn('Saved token expired or invalid:', err);
        clearSavedToken();
        setToken(null);
        setUser(null);
        setTenants([]);
        setActiveTenant(null);
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = async (email: string, pass: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.login({ email, password: pass });
      setSavedToken(res.token);
      setToken(res.token);
      setUser(res.user);
      setTenants(res.tenants);
      if (res.tenants.length > 0) {
        setActiveTenant(res.tenants[0]);
      }
    } catch (err: any) {
      setError(err.message || 'Authentication failed');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (payload: {
    email: string;
    password: string;
    full_name: string;
    phone_number?: string;
    company_name: string;
  }) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.register(payload);
      setSavedToken(res.token);
      setToken(res.token);
      setUser(res.user);
      setTenants(res.tenants);
      if (res.tenants.length > 0) {
        setActiveTenant(res.tenants[0]);
      }
    } catch (err: any) {
      setError(err.message || 'Registration failed');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await api.logout();
    } catch (err) {
      console.warn('Logout error ignored:', err);
    } finally {
      clearSavedToken();
      setToken(null);
      setUser(null);
      setTenants([]);
      setActiveTenant(null);
      setIsLoading(false);
    }
  };

  const clearError = () => setError(null);

  return (
    <AuthContext.Provider
      value={{
        user,
        tenants,
        activeTenant,
        token,
        isLoading,
        error,
        login,
        register,
        logout,
        setActiveTenant,
        clearError,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

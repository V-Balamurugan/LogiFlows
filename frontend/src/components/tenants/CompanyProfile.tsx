import React, { useState, useEffect, useCallback } from 'react';
import { 
  Building2, 
  ShieldCheck, 
  Mail, 
  Calendar, 
  Hash, 
  Edit3, 
  Save, 
  X, 
  Users, 
  AlertCircle, 
  CheckCircle2, 
  RefreshCw,
  Sparkles,
  Layers
} from 'lucide-react';
import { api } from '../../services/api';
import { TeamModal } from './TeamModal';
import type { TenantSummary } from '../../types/auth';

interface CompanyProfileProps {
  tenantId: string;
  userRole?: string;
  onTenantUpdated?: (updated: { name: string; contact_email?: string }) => void;
}

interface CompanyDetails {
  id: string;
  name: string;
  slug: string;
  status: string;
  contact_email: string;
  created_at?: string;
}

export const CompanyProfile: React.FC<CompanyProfileProps> = ({
  tenantId,
  userRole = 'VIEWER',
  onTenantUpdated,
}) => {
  const [company, setCompany] = useState<CompanyDetails | null>(null);
  const [branchCount, setBranchCount] = useState<number>(0);
  const [memberCount, setMemberCount] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Edit Form State
  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editEmail, setEditEmail] = useState('');
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState<string | null>(null);

  // Team Modal State
  const [isTeamModalOpen, setIsTeamModalOpen] = useState(false);

  const canEdit = userRole === 'TENANT_ADMIN' || userRole === 'PLATFORM_ADMIN';

  const fetchCompanyData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [tenantData, branchesRes, membersRes] = await Promise.all([
        api.getTenant(tenantId),
        api.listBranches(tenantId, { limit: 1 }).catch(() => ({ total: 0 })),
        api.listMembers(tenantId).catch(() => []),
      ]);

      setCompany(tenantData);
      setEditName(tenantData.name);
      setEditEmail(tenantData.contact_email || '');
      setBranchCount((branchesRes as any)?.total || 0);
      setMemberCount((membersRes as any)?.length || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to load company profile');
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    if (tenantId) {
      fetchCompanyData();
      setIsEditing(false);
      setSaveSuccess(null);
    }
  }, [tenantId, fetchCompanyData]);

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editName.trim()) {
      setError('Company name cannot be empty');
      return;
    }

    setSaving(true);
    setError(null);
    setSaveSuccess(null);

    try {
      const updated = await api.updateTenant(tenantId, {
        name: editName.trim(),
        contact_email: editEmail.trim() || undefined,
      });

      setCompany(prev => prev ? { ...prev, ...updated } : updated);
      setIsEditing(false);
      setSaveSuccess('Company profile successfully updated.');
      if (onTenantUpdated) {
        onTenantUpdated({ name: editName.trim(), contact_email: editEmail.trim() });
      }
    } catch (err: any) {
      setError(err.message || 'Failed to update company profile');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '5rem 0' }}>
        <RefreshCw size={36} className="animate-spin" style={{ color: 'var(--accent-cyan)', marginBottom: '1rem' }} />
        <p style={{ color: 'var(--text-muted)', fontSize: '0.95rem' }}>Loading company metadata & organization structure...</p>
      </div>
    );
  }

  if (error && !company) {
    return (
      <div style={{ padding: '2rem', background: 'rgba(239, 68, 68, 0.1)', border: '1px solid rgba(239, 68, 68, 0.3)', borderRadius: '12px', textAlign: 'center' }}>
        <AlertCircle size={40} style={{ color: '#ef4444', margin: '0 auto 1rem auto' }} />
        <h3 style={{ color: 'white', fontWeight: 600, fontSize: '1.1rem', marginBottom: '0.5rem' }}>Unable to Load Company Details</h3>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginBottom: '1.5rem' }}>{error}</p>
        <button
          onClick={fetchCompanyData}
          style={{
            padding: '8px 18px',
            background: 'var(--accent-cyan)',
            color: '#020617',
            fontWeight: 600,
            borderRadius: '6px',
            border: 'none',
            cursor: 'pointer'
          }}
        >
          Try Again
        </button>
      </div>
    );
  }

  if (!company) return null;

  const tenantSummary: TenantSummary = {
    id: company.id,
    name: company.name,
    slug: company.slug,
    role: userRole as any,
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Top Banner & Status */}
      <div style={{
        background: 'linear-gradient(135deg, rgba(30, 41, 59, 0.7) 0%, rgba(15, 23, 42, 0.9) 100%)',
        border: '1px solid var(--border-subtle)',
        borderRadius: '16px',
        padding: '2rem',
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '1rem' }}>
          <div style={{ display: 'flex', gap: '1.25rem', alignItems: 'center' }}>
            <div style={{
              width: '64px',
              height: '64px',
              borderRadius: '14px',
              background: 'linear-gradient(135deg, #0284c7 0%, #2563eb 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 4px 20px rgba(37, 99, 235, 0.4)'
            }}>
              <Building2 size={32} color="white" />
            </div>

            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.35rem' }}>
                <h1 style={{ fontSize: '1.6rem', fontWeight: 700, color: 'white', margin: 0 }}>
                  {company.name}
                </h1>
                <span style={{
                  padding: '4px 10px',
                  borderRadius: '20px',
                  fontSize: '0.75rem',
                  fontWeight: 700,
                  letterSpacing: '0.5px',
                  background: company.status === 'ACTIVE' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                  color: company.status === 'ACTIVE' ? '#34d399' : '#f87171',
                  border: `1px solid ${company.status === 'ACTIVE' ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`
                }}>
                  {company.status}
                </span>
              </div>
              <p style={{ color: 'var(--text-muted)', fontSize: '0.85rem', margin: 0 }}>
                Enterprise Identifier: <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--accent-cyan)' }}>{company.slug}</span>
              </p>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
            {canEdit && !isEditing && (
              <button
                onClick={() => setIsEditing(true)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  padding: '8px 16px',
                  background: 'rgba(59, 130, 246, 0.15)',
                  color: '#60a5fa',
                  border: '1px solid rgba(59, 130, 246, 0.3)',
                  borderRadius: '8px',
                  fontSize: '0.85rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                  transition: 'all 0.2s'
                }}
              >
                <Edit3 size={15} />
                Edit Details
              </button>
            )}

            <button
              onClick={() => setIsTeamModalOpen(true)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '8px 16px',
                background: 'linear-gradient(135deg, #0284c7 0%, #2563eb 100%)',
                color: 'white',
                border: 'none',
                borderRadius: '8px',
                fontSize: '0.85rem',
                fontWeight: 600,
                cursor: 'pointer',
                boxShadow: '0 2px 10px rgba(37, 99, 235, 0.3)'
              }}
            >
              <Users size={15} />
              Manage Team ({memberCount})
            </button>
          </div>
        </div>

        {/* Success Toast */}
        {saveSuccess && (
          <div style={{ marginTop: '1.25rem', padding: '10px 14px', background: 'rgba(16, 185, 129, 0.12)', border: '1px solid rgba(16, 185, 129, 0.3)', borderRadius: '8px', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#34d399', fontSize: '0.85rem' }}>
            <CheckCircle2 size={16} />
            {saveSuccess}
          </div>
        )}

        {/* Error Alert */}
        {error && (
          <div style={{ marginTop: '1.25rem', padding: '10px 14px', background: 'rgba(239, 68, 68, 0.12)', border: '1px solid rgba(239, 68, 68, 0.3)', borderRadius: '8px', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#f87171', fontSize: '0.85rem' }}>
            <AlertCircle size={16} />
            {error}
          </div>
        )}
      </div>

      {/* Edit Form Drawer / Card */}
      {isEditing && (
        <div style={{
          background: 'rgba(30, 41, 59, 0.95)',
          border: '1px solid rgba(59, 130, 246, 0.3)',
          borderRadius: '16px',
          padding: '1.75rem',
          boxShadow: '0 8px 30px rgba(0, 0, 0, 0.4)'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
            <h3 style={{ color: 'white', fontWeight: 600, fontSize: '1.1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Sparkles size={18} color="var(--accent-cyan)" />
              Update Organization Information
            </h3>
            <button
              onClick={() => { setIsEditing(false); setError(null); }}
              style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
            >
              <X size={18} />
            </button>
          </div>

          <form onSubmit={handleUpdate} style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.25rem' }}>
            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.4rem' }}>
                COMPANY NAME *
              </label>
              <input
                type="text"
                required
                value={editName}
                onChange={e => setEditName(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontSize: '0.9rem'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.4rem' }}>
                PRIMARY CONTACT EMAIL
              </label>
              <input
                type="email"
                value={editEmail}
                onChange={e => setEditEmail(e.target.value)}
                placeholder="ops@company.com"
                style={{
                  width: '100%',
                  padding: '10px 14px',
                  borderRadius: '8px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'white',
                  fontSize: '0.9rem'
                }}
              />
            </div>

            <div style={{ gridColumn: '1 / -1', display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
              <button
                type="button"
                onClick={() => setIsEditing(false)}
                style={{
                  padding: '8px 18px',
                  background: 'transparent',
                  border: '1px solid var(--border-subtle)',
                  color: 'var(--text-secondary)',
                  borderRadius: '8px',
                  cursor: 'pointer',
                  fontWeight: 500
                }}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  padding: '8px 20px',
                  background: 'var(--accent-cyan)',
                  border: 'none',
                  color: '#020617',
                  borderRadius: '8px',
                  cursor: saving ? 'not-allowed' : 'pointer',
                  fontWeight: 600,
                  opacity: saving ? 0.7 : 1
                }}
              >
                <Save size={16} />
                {saving ? 'Saving Changes...' : 'Save Profile'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Operational Metrics Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1.25rem' }}>
        <div style={{
          background: 'rgba(15, 23, 42, 0.6)',
          border: '1px solid var(--border-subtle)',
          borderRadius: '12px',
          padding: '1.25rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '0.5rem'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)', fontSize: '0.8rem', fontWeight: 600 }}>
            <span>DISTRIBUTION HUBS</span>
            <Layers size={16} color="var(--accent-cyan)" />
          </div>
          <div style={{ fontSize: '1.8rem', fontWeight: 700, color: 'white' }}>{branchCount}</div>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Registered active branches</span>
        </div>

        <div style={{
          background: 'rgba(15, 23, 42, 0.6)',
          border: '1px solid var(--border-subtle)',
          borderRadius: '12px',
          padding: '1.25rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '0.5rem'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)', fontSize: '0.8rem', fontWeight: 600 }}>
            <span>TEAM MEMBERS</span>
            <Users size={16} color="#60a5fa" />
          </div>
          <div style={{ fontSize: '1.8rem', fontWeight: 700, color: 'white' }}>{memberCount}</div>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Admins, operators, & viewers</span>
        </div>

        <div style={{
          background: 'rgba(15, 23, 42, 0.6)',
          border: '1px solid var(--border-subtle)',
          borderRadius: '12px',
          padding: '1.25rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '0.5rem'
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)', fontSize: '0.8rem', fontWeight: 600 }}>
            <span>YOUR PRIVILEGE ROLE</span>
            <ShieldCheck size={16} color="#34d399" />
          </div>
          <div style={{ fontSize: '1.2rem', fontWeight: 700, color: '#34d399', letterSpacing: '0.5px' }}>{userRole}</div>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Server-side verified access</span>
        </div>
      </div>

      {/* Metadata Overview Grid */}
      <div style={{
        background: 'rgba(15, 23, 42, 0.6)',
        border: '1px solid var(--border-subtle)',
        borderRadius: '14px',
        padding: '1.5rem'
      }}>
        <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'white', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Hash size={18} color="var(--accent-cyan)" />
          Enterprise Metadata & Compliance
        </h3>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.25rem' }}>
          <div>
            <span style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>ORGANIZATION UUID</span>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem', color: 'white', wordBreak: 'break-all' }}>{company.id}</span>
          </div>

          <div>
            <span style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>PRIMARY CONTACT EMAIL</span>
            <span style={{ fontSize: '0.85rem', color: 'white', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
              <Mail size={14} color="var(--text-muted)" />
              {company.contact_email || 'None registered'}
            </span>
          </div>

          <div>
            <span style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>ONBOARDING TIMESTAMP</span>
            <span style={{ fontSize: '0.85rem', color: 'white', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
              <Calendar size={14} color="var(--text-muted)" />
              {company.created_at ? new Date(company.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' }) : 'Phase 1 Genesis'}
            </span>
          </div>

          <div>
            <span style={{ display: 'block', fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>TENANT ISOLATION POLICIES</span>
            <span style={{ fontSize: '0.85rem', color: '#34d399', fontWeight: 600 }}>Strict Database & Schema Isolation Enforced</span>
          </div>
        </div>
      </div>

      {/* Team Management Modal */}
      {isTeamModalOpen && (
        <TeamModal
          tenant={tenantSummary}
          onClose={() => {
            setIsTeamModalOpen(false);
            fetchCompanyData();
          }}
        />
      )}
    </div>
  );
};

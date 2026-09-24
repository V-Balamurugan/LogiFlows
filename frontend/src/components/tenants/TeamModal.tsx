import React, { useState, useEffect } from 'react';
import { Users, X, UserPlus, Shield, Check, AlertCircle, Loader2 } from 'lucide-react';
import type { MemberDetails, TenantSummary } from '../../types/auth';
import { api } from '../../services/api';

interface TeamModalProps {
  tenant: TenantSummary;
  onClose: () => void;
}

export const TeamModal: React.FC<TeamModalProps> = ({ tenant, onClose }) => {
  const [members, setMembers] = useState<MemberDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form State
  const [newEmail, setNewEmail] = useState('');
  const [newRole, setNewRole] = useState<'TENANT_ADMIN' | 'TENANT_OPERATOR' | 'VIEWER'>('TENANT_OPERATOR');

  const fetchMembers = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.listMembers(tenant.id);
      setMembers(data);
    } catch (err: any) {
      setError(err.message || 'Failed to load team members');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMembers();
  }, [tenant.id]);

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setAdding(true);
    try {
      const added = await api.addMember(tenant.id, newEmail, newRole);
      setMembers((prev) => [...prev, added]);
      setSuccess(`User ${newEmail} added as ${newRole}`);
      setNewEmail('');
    } catch (err: any) {
      setError(err.message || 'Failed to add member');
    } finally {
      setAdding(false);
    }
  };

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'PLATFORM_ADMIN':
        return '#f43f5e';
      case 'TENANT_ADMIN':
        return '#3b82f6';
      case 'TENANT_OPERATOR':
        return '#10b981';
      default:
        return '#8b5cf6';
    }
  };

  return (
    <div style={{
      position: 'fixed',
      inset: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.75)',
      backdropFilter: 'blur(8px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
      padding: '1rem'
    }}>
      <div className="glass-panel" style={{
        width: '100%',
        maxWidth: '680px',
        maxHeight: '90vh',
        display: 'flex',
        flexDirection: 'column',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.8)',
        overflow: 'hidden'
      }}>
        {/* Header */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '1.25rem 1.5rem',
          borderBottom: '1px solid var(--border-subtle)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div style={{
              padding: '0.5rem',
              borderRadius: '8px',
              backgroundColor: 'rgba(6, 182, 212, 0.1)',
              color: 'var(--accent-cyan)'
            }}>
              <Users size={20} />
            </div>
            <div>
              <h2 style={{ fontSize: '1.15rem', fontWeight: 600 }}>Team & Role Management</h2>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                {tenant.name} ({tenant.slug})
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-muted)',
              cursor: 'pointer',
              padding: '0.25rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div style={{ padding: '1.5rem', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          {error && (
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.75rem 1rem',
              backgroundColor: 'rgba(244, 63, 94, 0.1)',
              border: '1px solid rgba(244, 63, 94, 0.3)',
              borderRadius: '8px',
              color: '#fda4af',
              fontSize: '0.85rem'
            }}>
              <AlertCircle size={16} />
              <span>{error}</span>
            </div>
          )}

          {success && (
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.75rem 1rem',
              backgroundColor: 'rgba(16, 185, 129, 0.1)',
              border: '1px solid rgba(16, 185, 129, 0.3)',
              borderRadius: '8px',
              color: '#6ee7b7',
              fontSize: '0.85rem'
            }}>
              <Check size={16} />
              <span>{success}</span>
            </div>
          )}

          {/* Add Member Form (For TENANT_ADMIN) */}
          {(tenant.role === 'TENANT_ADMIN' || tenant.role === 'PLATFORM_ADMIN') ? (
            <form onSubmit={handleAddMember} style={{
              padding: '1rem',
              backgroundColor: 'rgba(15, 23, 42, 0.5)',
              borderRadius: '10px',
              border: '1px solid var(--border-subtle)'
            }}>
              <h3 style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <UserPlus size={16} color="#06b6d4" />
                Add Member to {tenant.name}
              </h3>
              <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr auto', gap: '0.75rem' }}>
                <input
                  type="email"
                  required
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  placeholder="user@logistics.com"
                  style={{
                    padding: '0.6rem 0.85rem',
                    backgroundColor: 'rgba(10, 14, 23, 0.8)',
                    border: '1px solid var(--border-subtle)',
                    borderRadius: '8px',
                    color: 'var(--text-primary)',
                    fontSize: '0.85rem',
                    outline: 'none'
                  }}
                />
                <select
                  value={newRole}
                  onChange={(e: any) => setNewRole(e.target.value)}
                  style={{
                    padding: '0.6rem 0.85rem',
                    backgroundColor: 'rgba(10, 14, 23, 0.8)',
                    border: '1px solid var(--border-subtle)',
                    borderRadius: '8px',
                    color: 'var(--text-primary)',
                    fontSize: '0.85rem',
                    outline: 'none'
                  }}
                >
                  <option value="TENANT_ADMIN">Tenant Admin</option>
                  <option value="TENANT_OPERATOR">Tenant Operator</option>
                  <option value="VIEWER">Viewer (Read-only)</option>
                </select>
                <button
                  type="submit"
                  disabled={adding}
                  style={{
                    padding: '0.6rem 1rem',
                    backgroundColor: 'var(--accent-cyan)',
                    color: '#000',
                    border: 'none',
                    borderRadius: '8px',
                    fontWeight: 600,
                    fontSize: '0.85rem',
                    cursor: adding ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.35rem'
                  }}
                >
                  {adding ? <Loader2 size={16} className="spin-slow" /> : <UserPlus size={16} />}
                  <span>Add</span>
                </button>
              </div>
            </form>
          ) : (
            <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              Note: You have role <strong>{tenant.role}</strong>. Only Tenant Administrators can add members.
            </div>
          )}

          {/* Members Table */}
          <div>
            <h3 style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>
              Active Members ({members.length})
            </h3>
            {loading ? (
              <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>
                <Loader2 size={24} className="spin-slow" style={{ margin: '0 auto 0.5rem' }} />
                <span>Loading team members...</span>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                {members.map((m) => (
                  <div key={m.id} style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: '0.75rem 1rem',
                    backgroundColor: 'rgba(15, 23, 42, 0.4)',
                    border: '1px solid var(--border-subtle)',
                    borderRadius: '8px'
                  }}>
                    <div>
                      <div style={{ fontWeight: 600, fontSize: '0.9rem' }}>{m.full_name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{m.email}</div>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                      <span style={{
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        padding: '0.2rem 0.6rem',
                        borderRadius: '6px',
                        backgroundColor: `${getRoleBadgeColor(m.role)}20`,
                        color: getRoleBadgeColor(m.role),
                        border: `1px solid ${getRoleBadgeColor(m.role)}40`,
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.25rem'
                      }}>
                        <Shield size={12} />
                        {m.role}
                      </span>
                      <span style={{
                        fontSize: '0.7rem',
                        color: 'var(--text-muted)'
                      }}>
                        {new Date(m.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

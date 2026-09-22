import React, { useState, useEffect, useCallback } from 'react';
import { 
  Building2, 
  Plus, 
  MapPin, 
  Navigation, 
  Search, 
  AlertCircle, 
  CheckCircle2, 
  Loader2, 
  Trash2, 
  Globe2,
  SlidersHorizontal 
} from 'lucide-react';
import type { Branch, CreateBranchPayload } from '../../types/resources';
import { api } from '../../services/api';

interface BranchListProps {
  tenantId: string;
  userRole: string;
}

export const BranchList: React.FC<BranchListProps> = ({ tenantId, userRole }) => {
  const [branches, setBranches] = useState<Branch[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Search and filter state
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  // Create Modal state
  const [showModal, setShowModal] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const [formData, setFormData] = useState<CreateBranchPayload>({
    branch_code: '',
    name: '',
    address: '',
    city: '',
    state: '',
    postal_code: '',
    country: 'India',
    latitude: 28.6139,
    longitude: 77.2090,
    coverage_radius_km: 15,
  });

  const canManage = userRole === 'TENANT_ADMIN' || userRole === 'PLATFORM_ADMIN';

  const fetchBranches = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listBranches(tenantId, {
        search: searchTerm || undefined,
        status: statusFilter || undefined,
        limit: 50,
      });
      setBranches(res.branches || []);
      setTotal(res.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to load delivery branches');
    } finally {
      setLoading(false);
    }
  }, [tenantId, searchTerm, statusFilter]);

  useEffect(() => {
    fetchBranches();
  }, [fetchBranches]);

  const handleCreateBranch = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setCreating(true);

    if (formData.latitude < -90 || formData.latitude > 90 || formData.longitude < -180 || formData.longitude > 180) {
      setFormError('Latitude must be between -90 and 90, Longitude between -180 and 180');
      setCreating(false);
      return;
    }

    try {
      const newBranch = await api.createBranch(tenantId, {
        ...formData,
        branch_code: formData.branch_code.trim().toUpperCase(),
        latitude: Number(formData.latitude),
        longitude: Number(formData.longitude),
        coverage_radius_km: Number(formData.coverage_radius_km) || 15,
      });
      setBranches((prev) => [newBranch, ...prev]);
      setTotal((prev) => prev + 1);
      setSuccess(`Branch "${newBranch.name}" (${newBranch.branch_code}) created successfully`);
      setShowModal(false);
      setFormData({
        branch_code: '',
        name: '',
        address: '',
        city: '',
        state: '',
        postal_code: '',
        country: 'India',
        latitude: 28.6139,
        longitude: 77.2090,
        coverage_radius_km: 15,
      });
    } catch (err: any) {
      setFormError(err.message || 'Failed to create branch');
    } finally {
      setCreating(false);
    }
  };

  const handleDeactivate = async (branchId: string, branchName: string) => {
    if (!confirm(`Are you sure you want to deactivate branch "${branchName}"?`)) {
      return;
    }

    try {
      await api.deleteBranch(tenantId, branchId);
      setBranches((prev) =>
        prev.map((b) => (b.id === branchId ? { ...b, is_active: false, status: 'INACTIVE' } : b))
      );
      setSuccess(`Branch "${branchName}" deactivated successfully`);
    } catch (err: any) {
      setError(err.message || 'Failed to deactivate branch');
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Header & Controls */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: '1rem',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{
            width: '40px',
            height: '40px',
            borderRadius: '10px',
            background: 'rgba(56, 189, 248, 0.1)',
            border: '1px solid rgba(56, 189, 248, 0.25)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--accent-cyan)',
          }}>
            <Building2 size={22} />
          </div>
          <div>
            <h2 style={{ fontSize: '1.35rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
              Delivery Branches & Sorting Hubs
            </h2>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', margin: 0 }}>
              {total} registered logistics hubs with PostGIS spatial coverage
            </p>
          </div>
        </div>

        {canManage && (
          <button
            onClick={() => setShowModal(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.65rem 1.25rem',
              borderRadius: '8px',
              border: 'none',
              background: 'linear-gradient(135deg, #0284c7 0%, #0369a1 100%)',
              color: '#ffffff',
              fontWeight: 600,
              fontSize: '0.9rem',
              cursor: 'pointer',
              boxShadow: '0 4px 14px rgba(2, 132, 199, 0.35)',
              transition: 'all 0.2s ease',
            }}
          >
            <Plus size={18} />
            Add Delivery Branch
          </button>
        )}
      </div>

      {/* Notifications */}
      {error && (
        <div style={{
          padding: '0.85rem 1.25rem',
          borderRadius: '8px',
          background: 'rgba(244, 63, 94, 0.1)',
          border: '1px solid rgba(244, 63, 94, 0.3)',
          color: '#fda4af',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          fontSize: '0.9rem',
        }}>
          <AlertCircle size={18} />
          <span>{error}</span>
        </div>
      )}

      {success && (
        <div style={{
          padding: '0.85rem 1.25rem',
          borderRadius: '8px',
          background: 'rgba(16, 185, 129, 0.1)',
          border: '1px solid rgba(16, 185, 129, 0.3)',
          color: '#6ee7b7',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          fontSize: '0.9rem',
        }}>
          <CheckCircle2 size={18} />
          <span>{success}</span>
        </div>
      )}

      {/* Search & Filters */}
      <div style={{
        display: 'flex',
        gap: '1rem',
        flexWrap: 'wrap',
        alignItems: 'center',
        background: 'var(--bg-card)',
        padding: '0.85rem 1.25rem',
        borderRadius: '10px',
        border: '1px solid var(--border-subtle)',
      }}>
        <div style={{ position: 'relative', flex: '1', minWidth: '240px' }}>
          <Search size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
          <input
            type="text"
            placeholder="Search by branch code, name, city..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{
              width: '100%',
              padding: '0.55rem 0.85rem 0.55rem 2.25rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <SlidersHorizontal size={16} style={{ color: 'var(--text-muted)' }} />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.85rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Statuses</option>
            <option value="ACTIVE">Active</option>
            <option value="INACTIVE">Inactive</option>
          </select>
        </div>
      </div>

      {/* Branch Cards Grid */}
      {loading ? (
        <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          <Loader2 size={32} className="animate-spin" style={{ margin: '0 auto 1rem', display: 'block' }} />
          <p>Loading delivery branches...</p>
        </div>
      ) : branches.length === 0 ? (
        <div style={{
          padding: '4rem 2rem',
          textAlign: 'center',
          background: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px dashed var(--border-subtle)',
        }}>
          <Building2 size={48} style={{ color: 'var(--text-muted)', margin: '0 auto 1rem', opacity: 0.5 }} />
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
            No branches found
          </h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', maxWidth: '400px', margin: '0 auto 1.5rem' }}>
            {searchTerm ? 'No branches match your search criteria.' : 'Start coordinating deliveries by registering your first hub.'}
          </p>
          {canManage && !searchTerm && (
            <button
              onClick={() => setShowModal(true)}
              style={{
                padding: '0.65rem 1.25rem',
                borderRadius: '8px',
                border: 'none',
                background: 'var(--accent-cyan)',
                color: '#000',
                fontWeight: 600,
                cursor: 'pointer',
              }}
            >
              Add First Branch
            </button>
          )}
        </div>
      ) : (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
          gap: '1.25rem',
        }}>
          {branches.map((b) => (
            <div
              key={b.id}
              style={{
                background: 'var(--bg-card)',
                borderRadius: '12px',
                border: '1px solid var(--border-subtle)',
                padding: '1.25rem',
                display: 'flex',
                flexDirection: 'column',
                gap: '1rem',
                transition: 'all 0.2s ease',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <span style={{
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    letterSpacing: '0.05em',
                    padding: '0.2rem 0.5rem',
                    borderRadius: '4px',
                    background: 'rgba(56, 189, 248, 0.15)',
                    color: 'var(--accent-cyan)',
                    border: '1px solid rgba(56, 189, 248, 0.3)',
                    display: 'inline-block',
                    marginBottom: '0.5rem',
                  }}>
                    {b.branch_code}
                  </span>
                  <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
                    {b.name}
                  </h3>
                </div>

                <span style={{
                  fontSize: '0.7rem',
                  fontWeight: 600,
                  padding: '0.25rem 0.6rem',
                  borderRadius: '12px',
                  background: b.is_active ? 'rgba(16, 185, 129, 0.15)' : 'rgba(244, 63, 94, 0.15)',
                  color: b.is_active ? '#34d399' : '#fb7185',
                  border: `1px solid ${b.is_active ? 'rgba(16, 185, 129, 0.3)' : 'rgba(244, 63, 94, 0.3)'}`,
                }}>
                  {b.status}
                </span>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.5rem' }}>
                  <MapPin size={16} style={{ color: 'var(--accent-blue)', flexShrink: 0, marginTop: '2px' }} />
                  <span>{b.address}, {b.city}{b.state ? `, ${b.state}` : ''} {b.postal_code || ''}</span>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Navigation size={16} style={{ color: 'var(--accent-purple)', flexShrink: 0 }} />
                  <span>Lat: {b.latitude.toFixed(4)}, Lng: {b.longitude.toFixed(4)}</span>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Globe2 size={16} style={{ color: 'var(--accent-emerald)', flexShrink: 0 }} />
                  <span>Coverage: <strong>{b.coverage_radius_km} km</strong></span>
                  {b.distance_km !== undefined && b.distance_km !== null && (
                    <span style={{ color: 'var(--accent-amber)', marginLeft: 'auto' }}>
                      ({b.distance_km.toFixed(1)} km away)
                    </span>
                  )}
                </div>
              </div>

              {canManage && b.is_active && (
                <div style={{ marginTop: 'auto', paddingTop: '0.75rem', borderTop: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'flex-end' }}>
                  <button
                    onClick={() => handleDeactivate(b.id, b.name)}
                    style={{
                      background: 'none',
                      border: 'none',
                      color: 'var(--text-muted)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.4rem',
                      fontSize: '0.8rem',
                      cursor: 'pointer',
                      padding: '0.3rem 0.6rem',
                      borderRadius: '4px',
                    }}
                    onMouseEnter={(e) => (e.currentTarget.style.color = '#f43f5e')}
                    onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
                  >
                    <Trash2 size={14} />
                    Deactivate Hub
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Add Branch Modal */}
      {showModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(0, 0, 0, 0.75)',
          backdropFilter: 'blur(6px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '1rem',
        }}>
          <div style={{
            background: 'var(--bg-secondary)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '14px',
            width: '100%',
            maxWidth: '560px',
            maxHeight: '90vh',
            overflowY: 'auto',
            padding: '1.75rem',
            boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                <Building2 size={22} style={{ color: 'var(--accent-cyan)' }} />
                <h3 style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                  Register New Branch Hub
                </h3>
              </div>
              <button
                onClick={() => setShowModal(false)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '1.2rem' }}
              >
                ✕
              </button>
            </div>

            {formError && (
              <div style={{
                padding: '0.75rem 1rem',
                borderRadius: '8px',
                background: 'rgba(244, 63, 94, 0.1)',
                border: '1px solid rgba(244, 63, 94, 0.3)',
                color: '#fda4af',
                marginBottom: '1rem',
                fontSize: '0.85rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}>
                <AlertCircle size={16} />
                <span>{formError}</span>
              </div>
            )}

            <form onSubmit={handleCreateBranch} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Branch Code *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. DEL-NORTH"
                    value={formData.branch_code}
                    onChange={(e) => setFormData({ ...formData, branch_code: e.target.value })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Hub Name *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Delhi North Sorting Hub"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Street Address *
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Plot 12, Industrial Area, GT Karnal Rd"
                  value={formData.address}
                  onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                  style={{
                    width: '100%',
                    padding: '0.55rem 0.75rem',
                    borderRadius: '6px',
                    border: '1px solid var(--border-subtle)',
                    background: 'rgba(15, 23, 42, 0.8)',
                    color: 'var(--text-primary)',
                    fontSize: '0.85rem',
                  }}
                />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.75rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    City *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="City"
                    value={formData.city}
                    onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    State
                  </label>
                  <input
                    type="text"
                    placeholder="State"
                    value={formData.state}
                    onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    PIN / Postal
                  </label>
                  <input
                    type="text"
                    placeholder="110001"
                    value={formData.postal_code}
                    onChange={(e) => setFormData({ ...formData, postal_code: e.target.value })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0.75rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Latitude *
                  </label>
                  <input
                    type="number"
                    step="any"
                    required
                    placeholder="28.7041"
                    value={formData.latitude}
                    onChange={(e) => setFormData({ ...formData, latitude: parseFloat(e.target.value) || 0 })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Longitude *
                  </label>
                  <input
                    type="number"
                    step="any"
                    required
                    placeholder="77.1025"
                    value={formData.longitude}
                    onChange={(e) => setFormData({ ...formData, longitude: parseFloat(e.target.value) || 0 })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Coverage (km)
                  </label>
                  <input
                    type="number"
                    step="any"
                    min="1"
                    placeholder="15"
                    value={formData.coverage_radius_km}
                    onChange={(e) => setFormData({ ...formData, coverage_radius_km: parseFloat(e.target.value) || 15 })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  />
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1rem' }}>
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  style={{
                    padding: '0.65rem 1.25rem',
                    borderRadius: '8px',
                    border: '1px solid var(--border-subtle)',
                    background: 'transparent',
                    color: 'var(--text-secondary)',
                    fontWeight: 600,
                    cursor: 'pointer',
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={creating}
                  style={{
                    padding: '0.65rem 1.5rem',
                    borderRadius: '8px',
                    border: 'none',
                    background: 'linear-gradient(135deg, #0284c7 0%, #0369a1 100%)',
                    color: '#ffffff',
                    fontWeight: 600,
                    cursor: creating ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                  }}
                >
                  {creating ? <Loader2 size={16} className="animate-spin" /> : null}
                  {creating ? 'Saving...' : 'Register Branch'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

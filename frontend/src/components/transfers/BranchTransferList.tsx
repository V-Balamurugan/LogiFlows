import React, { useState, useEffect } from 'react';
import { 
  ArrowLeftRight, 
  Plus, 
  CheckCircle2, 
  AlertCircle, 
  RefreshCw, 
  X,
  PackageCheck,
  Send,
  ArrowRight
} from 'lucide-react';
import { api } from '../../services/api';
import type { 
  BranchTransfer, 
  TransferStatus, 
  Parcel
} from '../../types/parcels';
import type { Branch, Employee, Vehicle } from '../../types/resources';

interface BranchTransferListProps {
  tenantId: string;
  userRole?: string;
}

const TRANSFER_STATUS_COLORS: Record<TransferStatus, { bg: string; text: string; border: string }> = {
  INITIATED: { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)' },
  IN_TRANSIT: { bg: 'rgba(139, 92, 246, 0.2)', text: '#a78bfa', border: 'rgba(139, 92, 246, 0.4)' },
  RECEIVED: { bg: 'rgba(16, 185, 129, 0.18)', text: '#34d399', border: 'rgba(16, 185, 129, 0.4)' },
  CANCELLED: { bg: 'rgba(239, 68, 68, 0.15)', text: '#ef4444', border: 'rgba(239, 68, 68, 0.3)' },
};

export const BranchTransferList: React.FC<BranchTransferListProps> = ({ tenantId, userRole }) => {
  const [transfers, setTransfers] = useState<BranchTransfer[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [parcels, setParcels] = useState<Parcel[]>([]);
  const [drivers, setDrivers] = useState<Employee[]>([]);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [selectedTransfer, setSelectedTransfer] = useState<BranchTransfer | null>(null);

  // Form state
  const [originBranchId, setOriginBranchId] = useState('');
  const [destBranchId, setDestBranchId] = useState('');
  const [driverId, setDriverId] = useState('');
  const [vehicleId, setVehicleId] = useState('');
  const [selectedParcelIds, setSelectedParcelIds] = useState<string[]>([]);
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const isOperator = userRole === 'TENANT_ADMIN' || userRole === 'TENANT_OPERATOR' || userRole === 'PLATFORM_ADMIN';

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [transRes, branchRes, parcRes, empRes, vehRes] = await Promise.all([
        api.listBranchTransfers(tenantId),
        api.listBranches(tenantId),
        api.listParcels(tenantId, { limit: 100 }),
        api.listEmployees(tenantId, { operational_role: 'DRIVER' }),
        api.listVehicles(tenantId),
      ]);

      setTransfers(transRes.transfers || []);
      setBranches(branchRes.branches || []);
      setParcels(parcRes.parcels || []);
      setDrivers(empRes.employees || []);
      setVehicles(vehRes.vehicles || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load branch transfers');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [tenantId]);

  // Filter parcels located at the selected origin hub branch that are ready for transfer
  const availableParcels = parcels.filter(
    (p) => p.current_branch_id === originBranchId && p.status !== 'DELIVERED' && p.status !== 'CANCELLED' && p.status !== 'IN_TRANSIT'
  );

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (originBranchId === destBranchId) {
      setFormError('Source and destination branches must be distinct');
      return;
    }
    if (selectedParcelIds.length === 0) {
      setFormError('Please select at least 1 parcel to include in this linehaul transfer manifest');
      return;
    }

    try {
      setSubmitting(true);
      setFormError(null);
      await api.createBranchTransfer(tenantId, {
        origin_branch_id: originBranchId,
        destination_branch_id: destBranchId,
        responsible_employee_id: driverId || undefined,
        vehicle_id: vehicleId || undefined,
        parcel_ids: selectedParcelIds,
        notes: notes || undefined,
      });

      setShowCreateModal(false);
      setSuccessMsg('Transfer manifest generated successfully!');
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      setFormError(err.message || 'Failed to generate transfer manifest');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDispatch = async (t: BranchTransfer) => {
    if (!confirm(`Dispatch linehaul transfer ${t.transfer_number} to ${t.destination_branch_name}? Parcels will transition to IN_TRANSIT.`)) {
      return;
    }
    try {
      await api.dispatchBranchTransfer(tenantId, t.id, 'Dispatched from origin hub terminal');
      setSuccessMsg(`Transfer ${t.transfer_number} is now IN_TRANSIT!`);
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      alert(`Dispatch failed: ${err.message}`);
    }
  };

  const handleReceive = async (t: BranchTransfer) => {
    if (!confirm(`Confirm arrival of cargo manifest ${t.transfer_number} at ${t.destination_branch_name}?`)) {
      return;
    }
    try {
      await api.receiveBranchTransfer(tenantId, t.id, undefined, 'Received and verified at destination sorting bay');
      setSuccessMsg(`Transfer ${t.transfer_number} received! Parcels checked in at hub.`);
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      alert(`Receive operation failed: ${err.message}`);
    }
  };

  const handleOpenDetail = async (t: BranchTransfer) => {
    try {
      const full = await api.getBranchTransfer(tenantId, t.id);
      setSelectedTransfer(full);
      setShowDetailModal(true);
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <ArrowLeftRight size={26} color="var(--accent-purple)" />
            Inter-Branch Linehaul Transfers
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginTop: '0.2rem' }}>
            Hub-to-hub cargo manifests, cross-dock chain-of-custody handovers, and in-transit tracking.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          {isOperator && (
            <button
              onClick={() => {
                setFormError(null);
                setOriginBranchId(branches.length > 0 ? branches[0].id : '');
                setDestBranchId(branches.length > 1 ? branches[1].id : '');
                setSelectedParcelIds([]);
                setNotes('');
                setShowCreateModal(true);
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.6rem 1.2rem',
                borderRadius: '8px',
                border: 'none',
                background: 'linear-gradient(135deg, var(--accent-purple) 0%, #6366f1 100%)',
                color: '#fff',
                cursor: 'pointer',
                fontWeight: 600,
                boxShadow: '0 4px 12px rgba(139, 92, 246, 0.25)',
              }}
            >
              <Plus size={18} />
              Create Transfer Manifest
            </button>
          )}

          <button
            onClick={loadData}
            style={{
              padding: '0.6rem',
              borderRadius: '8px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(30, 41, 59, 0.7)',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
            }}
            title="Refresh"
          >
            <RefreshCw size={17} />
          </button>
        </div>
      </div>

      {successMsg && (
        <div style={{
          padding: '0.8rem 1.2rem',
          borderRadius: '8px',
          background: 'rgba(16, 185, 129, 0.15)',
          border: '1px solid rgba(16, 185, 129, 0.3)',
          color: 'var(--accent-emerald)',
          display: 'flex',
          alignItems: 'center',
          gap: '0.6rem',
          fontSize: '0.9rem',
        }}>
          <CheckCircle2 size={18} />
          {successMsg}
        </div>
      )}

      {/* Main Table */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
          <RefreshCw className="spin" size={24} style={{ marginBottom: '0.5rem' }} />
          <p>Loading linehaul manifests...</p>
        </div>
      ) : error ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--accent-rose)' }}>
          <AlertCircle size={28} style={{ margin: '0 auto 0.5rem auto' }} />
          <p>{error}</p>
        </div>
      ) : transfers.length === 0 ? (
        <div className="glass-panel" style={{ padding: '4rem 2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          <ArrowLeftRight size={44} style={{ margin: '0 auto 1rem auto', opacity: 0.4 }} />
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-secondary)' }}>No branch transfers on record</h3>
          <p style={{ fontSize: '0.85rem', marginTop: '0.25rem' }}>
            Click "Create Transfer Manifest" to route bulk parcels between sorting hubs.
          </p>
        </div>
      ) : (
        <div className="glass-panel" style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem', textAlign: 'left' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>
                <th style={{ padding: '1rem' }}>Manifest #</th>
                <th style={{ padding: '1rem' }}>Route (Source &rarr; Target Hub)</th>
                <th style={{ padding: '1rem' }}>Cargo Quantity</th>
                <th style={{ padding: '1rem' }}>Linehaul Driver / Vehicle</th>
                <th style={{ padding: '1rem' }}>Transfer Status</th>
                <th style={{ padding: '1rem', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {transfers.map((t) => {
                const st = TRANSFER_STATUS_COLORS[t.status] || { bg: 'rgba(255,255,255,0.1)', text: '#fff', border: 'transparent' };
                return (
                  <tr key={t.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                    <td style={{ padding: '1rem' }}>
                      <div style={{ fontWeight: 700, fontFamily: 'monospace', color: 'var(--accent-purple)' }}>
                        {t.transfer_number}
                      </div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                        Created: {new Date(t.created_at).toLocaleDateString()}
                      </div>
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 600 }}>
                        <span>{t.origin_branch_name}</span>
                        <ArrowRight size={13} color="var(--text-muted)" />
                        <span style={{ color: 'var(--accent-cyan)' }}>{t.destination_branch_name}</span>
                      </div>
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <span style={{
                        display: 'inline-block',
                        padding: '2px 8px',
                        borderRadius: '4px',
                        background: 'rgba(255,255,255,0.08)',
                        fontSize: '0.85rem',
                        fontWeight: 600
                      }}>
                        {t.parcels_count || (t.parcels ? t.parcels.length : 0)} packages
                      </span>
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <div style={{ fontSize: '0.85rem' }}>{t.employee_name || 'Unassigned Courier'}</div>
                      {t.vehicle_license_plate && (
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Vehicle: {t.vehicle_license_plate}</div>
                      )}
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <span style={{
                        display: 'inline-block',
                        padding: '4px 10px',
                        borderRadius: '20px',
                        fontSize: '0.75rem',
                        fontWeight: 700,
                        background: st.bg,
                        color: st.text,
                        border: `1px solid ${st.border}`
                      }}>
                        {t.status.replace(/_/g, ' ')}
                      </span>
                    </td>

                    <td style={{ padding: '1rem', textAlign: 'right' }}>
                      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.4rem' }}>
                        <button
                          onClick={() => handleOpenDetail(t)}
                          style={{
                            padding: '0.4rem 0.75rem',
                            borderRadius: '6px',
                            border: '1px solid var(--border-subtle)',
                            background: 'rgba(30, 41, 59, 0.7)',
                            color: 'var(--text-primary)',
                            fontSize: '0.8rem',
                            cursor: 'pointer'
                          }}
                        >
                          Manifest
                        </button>

                        {isOperator && t.status === 'INITIATED' && (
                          <button
                            onClick={() => handleDispatch(t)}
                            style={{
                              padding: '0.4rem 0.75rem',
                              borderRadius: '6px',
                              border: 'none',
                              background: 'rgba(139, 92, 246, 0.25)',
                              color: '#a78bfa',
                              fontSize: '0.8rem',
                              fontWeight: 600,
                              cursor: 'pointer',
                              display: 'flex',
                              alignItems: 'center',
                              gap: '4px'
                            }}
                          >
                            <Send size={13} />
                            Dispatch
                          </button>
                        )}

                        {isOperator && t.status === 'IN_TRANSIT' && (
                          <button
                            onClick={() => handleReceive(t)}
                            style={{
                              padding: '0.4rem 0.75rem',
                              borderRadius: '6px',
                              border: 'none',
                              background: 'rgba(16, 185, 129, 0.25)',
                              color: '#34d399',
                              fontSize: '0.8rem',
                              fontWeight: 600,
                              cursor: 'pointer',
                              display: 'flex',
                              alignItems: 'center',
                              gap: '4px'
                            }}
                          >
                            <PackageCheck size={14} />
                            Receive Hub
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* CREATE TRANSFER MANIFEST MODAL */}
      {showCreateModal && (
        <div style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0, 0, 0, 0.75)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '1rem',
          backdropFilter: 'blur(5px)',
        }}>
          <div className="glass-panel" style={{ width: '100%', maxWidth: '640px', maxHeight: '90vh', overflowY: 'auto', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <ArrowLeftRight size={22} color="var(--accent-purple)" />
                Assemble Inter-Branch Cargo Manifest
              </h3>
              <button onClick={() => setShowCreateModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            {formError && (
              <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--accent-rose)', color: 'var(--accent-rose)', marginBottom: '1rem', fontSize: '0.85rem' }}>
                {formError}
              </div>
            )}

            <form onSubmit={handleCreateSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1.2rem' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Source Origin Hub *</label>
                  <select
                    required
                    value={originBranchId}
                    onChange={(e) => {
                      setOriginBranchId(e.target.value);
                      setSelectedParcelIds([]);
                    }}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">Select origin hub...</option>
                    {branches.map((b) => (
                      <option key={b.id} value={b.id}>{b.name} ({b.city})</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Target Destination Hub *</label>
                  <select
                    required
                    value={destBranchId}
                    onChange={(e) => setDestBranchId(e.target.value)}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">Select target hub...</option>
                    {branches.map((b) => (
                      <option key={b.id} value={b.id}>{b.name} ({b.city})</option>
                    ))}
                  </select>
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Linehaul Driver</label>
                  <select
                    value={driverId}
                    onChange={(e) => setDriverId(e.target.value)}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">None / External Logistics</option>
                    {drivers.map((d) => (
                      <option key={d.id} value={d.id}>{d.first_name} {d.last_name}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Transport Vehicle</label>
                  <select
                    value={vehicleId}
                    onChange={(e) => setVehicleId(e.target.value)}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">None</option>
                    {vehicles.map((v) => (
                      <option key={v.id} value={v.id}>{v.registration_number} ({v.make_model || v.vehicle_type})</option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Parcel selection list */}
              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '6px' }}>
                  Select Packages to Manifest ({selectedParcelIds.length} selected)
                </label>
                {availableParcels.length === 0 ? (
                  <div style={{ padding: '1rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.5)', color: 'var(--text-muted)', fontSize: '0.85rem', textAlign: 'center' }}>
                    No packages currently located at the selected source hub branch.
                  </div>
                ) : (
                  <div style={{ maxHeight: '180px', overflowY: 'auto', border: '1px solid var(--border-subtle)', borderRadius: '6px', padding: '0.5rem', background: 'rgba(15, 23, 42, 0.5)' }}>
                    {availableParcels.map((p) => {
                      const isChecked = selectedParcelIds.includes(p.id);
                      return (
                        <div
                          key={p.id}
                          onClick={() => {
                            if (isChecked) {
                              setSelectedParcelIds(selectedParcelIds.filter((id) => id !== p.id));
                            } else {
                              setSelectedParcelIds([...selectedParcelIds, p.id]);
                            }
                          }}
                          style={{
                            padding: '0.4rem 0.6rem',
                            borderRadius: '4px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.6rem',
                            cursor: 'pointer',
                            background: isChecked ? 'rgba(139, 92, 246, 0.15)' : 'transparent',
                            fontSize: '0.85rem'
                          }}
                        >
                          <input type="checkbox" checked={isChecked} readOnly />
                          <span style={{ fontWeight: 600, fontFamily: 'monospace', color: 'var(--accent-cyan)' }}>{p.tracking_number}</span>
                          <span style={{ color: 'var(--text-muted)' }}>&rarr; {p.destination_branch_name} ({p.weight_kg} kg)</span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>

              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Manifest Notes</label>
                <input
                  type="text"
                  placeholder="e.g. Scheduled Night Inter-City Shuttle #4"
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1rem' }}>
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  style={{ padding: '0.6rem 1.2rem', borderRadius: '6px', border: '1px solid var(--border-subtle)', background: 'transparent', color: 'var(--text-secondary)', cursor: 'pointer' }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting || selectedParcelIds.length === 0}
                  style={{ padding: '0.6rem 1.4rem', borderRadius: '6px', border: 'none', background: 'var(--accent-purple)', color: '#fff', fontWeight: 600, cursor: 'pointer' }}
                >
                  {submitting ? 'Generating Manifest...' : 'Issue Transfer Manifest'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* DETAIL MODAL */}
      {showDetailModal && selectedTransfer && (
        <div style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0, 0, 0, 0.75)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '1rem',
          backdropFilter: 'blur(5px)',
        }}>
          <div className="glass-panel" style={{ width: '100%', maxWidth: '640px', maxHeight: '90vh', overflowY: 'auto', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '1rem' }}>
              <div>
                <span style={{ fontSize: '1.3rem', fontWeight: 700, fontFamily: 'monospace', color: 'var(--accent-purple)' }}>
                  {selectedTransfer.transfer_number}
                </span>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                  Route: {selectedTransfer.origin_branch_name} &rarr; {selectedTransfer.destination_branch_name}
                </div>
              </div>
              <button onClick={() => setShowDetailModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            <h4 style={{ fontSize: '0.95rem', fontWeight: 600, marginBottom: '0.75rem' }}>Manifest Cargo Items</h4>
            {(!selectedTransfer.parcels || selectedTransfer.parcels.length === 0) ? (
              <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>No packages associated with this transfer.</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                {selectedTransfer.parcels.map((p) => (
                  <div key={p.id} style={{ padding: '0.6rem 0.8rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.6)', border: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                      <span style={{ fontFamily: 'monospace', fontWeight: 700, color: 'var(--accent-cyan)' }}>{p.tracking_number}</span>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>To: {p.receiver_name} ({p.receiver_address})</div>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                      {p.weight_kg} kg • {p.status}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

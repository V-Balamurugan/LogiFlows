import React, { useState, useEffect, useCallback } from 'react';
import { 
  Truck, 
  Plus, 
  Search, 
  AlertCircle, 
  CheckCircle2, 
  Loader2, 
  Building2, 
  UserCheck, 
  UserX, 
  Trash2,
  SlidersHorizontal,
  Weight,
  Box,
  BatteryCharging
} from 'lucide-react';
import type { Vehicle, CreateVehiclePayload, Branch, Employee } from '../../types/resources';
import { api } from '../../services/api';

interface VehicleListProps {
  tenantId: string;
  userRole: string;
}

export const VehicleList: React.FC<VehicleListProps> = ({ tenantId, userRole }) => {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [drivers, setDrivers] = useState<Employee[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Search & Filters
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [branchFilter, setBranchFilter] = useState('');

  // Create Modal State
  const [showModal, setShowModal] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const [formData, setFormData] = useState<CreateVehiclePayload>({
    registration_number: '',
    vehicle_type: 'ELECTRIC_VAN',
    make_model: '',
    year: new Date().getFullYear(),
    max_weight_kg: 750,
    max_volume_cbm: 4.5,
    assigned_branch_id: '',
  });

  // Assign Modal State
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedVehicle, setSelectedVehicle] = useState<Vehicle | null>(null);
  const [selectedDriverId, setSelectedDriverId] = useState('');
  const [assigning, setAssigning] = useState(false);
  const [assignError, setAssignError] = useState<string | null>(null);

  const canManage = userRole === 'TENANT_ADMIN' || userRole === 'PLATFORM_ADMIN';
  const canAssign = canManage || userRole === 'TENANT_OPERATOR';

  // Fetch Branches and Drivers
  useEffect(() => {
    let isMounted = true;
    api.listBranches(tenantId, { limit: 100 })
      .then((res) => {
        if (isMounted) setBranches(res.branches || []);
      })
      .catch(() => {});

    api.listEmployees(tenantId, { operational_role: 'DRIVER', status: 'ACTIVE', limit: 100 })
      .then((res) => {
        if (isMounted) setDrivers(res.employees || []);
      })
      .catch(() => {});

    return () => {
      isMounted = false;
    };
  }, [tenantId]);

  const fetchVehicles = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listVehicles(tenantId, {
        search: searchTerm || undefined,
        vehicle_type: typeFilter || undefined,
        branch_id: branchFilter || undefined,
        status: statusFilter || undefined,
        limit: 50,
      });
      setVehicles(res.vehicles || []);
      setTotal(res.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to load fleet vehicles');
    } finally {
      setLoading(false);
    }
  }, [tenantId, searchTerm, typeFilter, branchFilter, statusFilter]);

  useEffect(() => {
    fetchVehicles();
  }, [fetchVehicles]);

  const handleCreateVehicle = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setCreating(true);

    try {
      const newV = await api.createVehicle(tenantId, {
        ...formData,
        registration_number: formData.registration_number.trim().toUpperCase(),
        make_model: formData.make_model?.trim() || undefined,
        year: Number(formData.year) || undefined,
        max_weight_kg: Number(formData.max_weight_kg) || 500,
        max_volume_cbm: Number(formData.max_volume_cbm) || 3.0,
        assigned_branch_id: formData.assigned_branch_id || undefined,
      });
      setVehicles((prev) => [newV, ...prev]);
      setTotal((prev) => prev + 1);
      setSuccess(`Vehicle "${newV.registration_number}" registered successfully`);
      setShowModal(false);
      setFormData({
        registration_number: '',
        vehicle_type: 'ELECTRIC_VAN',
        make_model: '',
        year: new Date().getFullYear(),
        max_weight_kg: 750,
        max_volume_cbm: 4.5,
        assigned_branch_id: '',
      });
    } catch (err: any) {
      setFormError(err.message || 'Failed to register vehicle');
    } finally {
      setCreating(false);
    }
  };

  const handleAssignDriver = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedVehicle || !selectedDriverId) return;

    setAssignError(null);
    setAssigning(true);

    try {
      await api.assignVehicle(tenantId, selectedVehicle.id, selectedDriverId);
      const assignedDriver = drivers.find((d) => d.id === selectedDriverId);
      const driverName = assignedDriver ? `${assignedDriver.first_name} ${assignedDriver.last_name}` : 'Driver';

      setVehicles((prev) =>
        prev.map((v) =>
          v.id === selectedVehicle.id
            ? { ...v, status: 'ASSIGNED', current_driver_id: selectedDriverId, current_driver_name: driverName }
            : v
        )
      );
      setSuccess(`Driver "${driverName}" assigned to vehicle "${selectedVehicle.registration_number}"`);
      setShowAssignModal(false);
      setSelectedVehicle(null);
      setSelectedDriverId('');
    } catch (err: any) {
      setAssignError(err.message || 'Failed to assign driver');
    } finally {
      setAssigning(false);
    }
  };

  const handleUnassignDriver = async (vehicle: Vehicle) => {
    if (!confirm(`Unassign current driver from vehicle "${vehicle.registration_number}"?`)) {
      return;
    }

    try {
      await api.unassignVehicle(tenantId, vehicle.id);
      setVehicles((prev) =>
        prev.map((v) =>
          v.id === vehicle.id
            ? { ...v, status: 'AVAILABLE', current_driver_id: undefined, current_driver_name: undefined }
            : v
        )
      );
      setSuccess(`Driver unassigned from vehicle "${vehicle.registration_number}"`);
    } catch (err: any) {
      setError(err.message || 'Failed to unassign driver');
    }
  };

  const handleDecommission = async (vehicleId: string, regNum: string) => {
    if (!confirm(`Are you sure you want to decommission vehicle "${regNum}"?`)) {
      return;
    }

    try {
      await api.deleteVehicle(tenantId, vehicleId);
      setVehicles((prev) =>
        prev.map((v) =>
          v.id === vehicleId
            ? { ...v, is_active: false, status: 'DECOMMISSIONED', current_driver_id: undefined, current_driver_name: undefined }
            : v
        )
      );
      setSuccess(`Vehicle "${regNum}" decommissioned`);
    } catch (err: any) {
      setError(err.message || 'Failed to decommission vehicle');
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'AVAILABLE':
        return { bg: 'rgba(16, 185, 129, 0.15)', text: '#34d399', border: 'rgba(16, 185, 129, 0.3)' };
      case 'ASSIGNED':
        return { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)' };
      case 'IN_TRANSIT':
        return { bg: 'rgba(245, 158, 11, 0.15)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.3)' };
      case 'MAINTENANCE':
        return { bg: 'rgba(244, 63, 94, 0.15)', text: '#fda4af', border: 'rgba(244, 63, 94, 0.3)' };
      case 'DECOMMISSIONED':
      default:
        return { bg: 'rgba(148, 163, 184, 0.15)', text: '#94a3b8', border: 'rgba(148, 163, 184, 0.3)' };
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
            background: 'rgba(6, 182, 212, 0.1)',
            border: '1px solid rgba(6, 182, 212, 0.25)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--accent-cyan)',
          }}>
            <Truck size={22} />
          </div>
          <div>
            <h2 style={{ fontSize: '1.35rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
              Fleet Vehicles & Asset Assignment
            </h2>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', margin: 0 }}>
              {total} registered delivery vehicles with driver scheduling & conflict resolution
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
            Register Vehicle
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
        <div style={{ position: 'relative', flex: '1', minWidth: '220px' }}>
          <Search size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
          <input
            type="text"
            placeholder="Search by registration number, make/model..."
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
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.85rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Vehicle Types</option>
            <option value="ELECTRIC_VAN">Electric Vans</option>
            <option value="VAN">Standard Vans</option>
            <option value="MOTORCYCLE">Motorcycles</option>
            <option value="TRUCK">Heavy Trucks</option>
            <option value="THREE_WHEELER">Three Wheelers</option>
          </select>

          <select
            value={branchFilter}
            onChange={(e) => setBranchFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.85rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Branches</option>
            {branches.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name} ({b.branch_code})
              </option>
            ))}
          </select>

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
            <option value="AVAILABLE">Available</option>
            <option value="ASSIGNED">Assigned</option>
            <option value="IN_TRANSIT">In Transit</option>
            <option value="MAINTENANCE">Maintenance</option>
            <option value="DECOMMISSIONED">Decommissioned</option>
          </select>
        </div>
      </div>

      {/* Vehicle Grid */}
      {loading ? (
        <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          <Loader2 size={32} className="animate-spin" style={{ margin: '0 auto 1rem', display: 'block' }} />
          <p>Loading fleet vehicles...</p>
        </div>
      ) : vehicles.length === 0 ? (
        <div style={{
          padding: '4rem 2rem',
          textAlign: 'center',
          background: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px dashed var(--border-subtle)',
        }}>
          <Truck size={48} style={{ color: 'var(--text-muted)', margin: '0 auto 1rem', opacity: 0.5 }} />
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
            No vehicles registered
          </h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', maxWidth: '400px', margin: '0 auto 1.5rem' }}>
            {searchTerm ? 'No fleet vehicles match your filter criteria.' : 'Register your delivery vans, trucks, and electric vehicles to begin route assignments.'}
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
              Register First Vehicle
            </button>
          )}
        </div>
      ) : (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
          gap: '1.25rem',
        }}>
          {vehicles.map((v) => {
            const statusStyle = getStatusBadge(v.status);
            return (
              <div
                key={v.id}
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
                      fontSize: '0.8rem',
                      fontWeight: 700,
                      letterSpacing: '0.05em',
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      background: 'rgba(56, 189, 248, 0.15)',
                      color: 'var(--accent-cyan)',
                      border: '1px solid rgba(56, 189, 248, 0.3)',
                      display: 'inline-block',
                      marginBottom: '0.4rem',
                    }}>
                      {v.registration_number}
                    </span>
                    <h3 style={{ fontSize: '1.05rem', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
                      {v.make_model || v.vehicle_type}
                    </h3>
                    {v.year && (
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        Model Year: {v.year}
                      </span>
                    )}
                  </div>

                  <span style={{
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    padding: '0.25rem 0.6rem',
                    borderRadius: '12px',
                    background: statusStyle.bg,
                    color: statusStyle.text,
                    border: `1px solid ${statusStyle.border}`,
                  }}>
                    {v.status}
                  </span>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.45rem', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                  {v.vehicle_type === 'ELECTRIC_VAN' && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-cyan)' }}>
                      <BatteryCharging size={16} />
                      <span>Zero-Emission Electric Vehicle</span>
                    </div>
                  )}

                  {v.branch_name && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <Building2 size={16} style={{ color: 'var(--accent-blue)', flexShrink: 0 }} />
                      <span>{v.branch_name} {v.branch_code ? `(${v.branch_code})` : ''}</span>
                    </div>
                  )}

                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.5rem', fontSize: '0.8rem', background: 'rgba(255, 255, 255, 0.03)', padding: '0.5rem 0.75rem', borderRadius: '6px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                      <Weight size={14} style={{ color: 'var(--text-muted)' }} />
                      <span>{v.max_weight_kg} kg</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                      <Box size={14} style={{ color: 'var(--text-muted)' }} />
                      <span>{v.max_volume_cbm} m³</span>
                    </div>
                  </div>

                  {/* Driver Assignment Status */}
                  <div style={{
                    marginTop: '0.25rem',
                    padding: '0.65rem 0.85rem',
                    borderRadius: '8px',
                    background: v.current_driver_name ? 'rgba(59, 130, 246, 0.08)' : 'rgba(255, 255, 255, 0.02)',
                    border: `1px solid ${v.current_driver_name ? 'rgba(59, 130, 246, 0.2)' : 'var(--border-subtle)'}`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <UserCheck size={16} style={{ color: v.current_driver_name ? 'var(--accent-blue)' : 'var(--text-muted)' }} />
                      <span style={{ fontSize: '0.8rem', color: v.current_driver_name ? 'var(--text-primary)' : 'var(--text-muted)', fontWeight: v.current_driver_name ? 600 : 400 }}>
                        {v.current_driver_name ? v.current_driver_name : 'No Active Driver'}
                      </span>
                    </div>

                    {canAssign && v.is_active && (
                      v.current_driver_name ? (
                        <button
                          onClick={() => handleUnassignDriver(v)}
                          title="Unassign Driver"
                          style={{
                            background: 'none',
                            border: 'none',
                            color: 'var(--text-muted)',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '0.25rem',
                            fontSize: '0.75rem',
                          }}
                          onMouseEnter={(e) => (e.currentTarget.style.color = '#f43f5e')}
                          onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
                        >
                          <UserX size={14} />
                          Unassign
                        </button>
                      ) : (
                        <button
                          onClick={() => {
                            setSelectedVehicle(v);
                            setShowAssignModal(true);
                          }}
                          style={{
                            background: 'rgba(59, 130, 246, 0.15)',
                            border: '1px solid rgba(59, 130, 246, 0.3)',
                            color: 'var(--accent-blue)',
                            padding: '0.25rem 0.5rem',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontSize: '0.75rem',
                            fontWeight: 600,
                          }}
                        >
                          Assign Driver
                        </button>
                      )
                    )}
                  </div>
                </div>

                {canManage && v.is_active && (
                  <div style={{ marginTop: 'auto', paddingTop: '0.75rem', borderTop: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'flex-end' }}>
                    <button
                      onClick={() => handleDecommission(v.id, v.registration_number)}
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
                      Decommission
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Register Vehicle Modal */}
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
            maxWidth: '540px',
            maxHeight: '90vh',
            overflowY: 'auto',
            padding: '1.75rem',
            boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                <Truck size={22} style={{ color: 'var(--accent-cyan)' }} />
                <h3 style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                  Register Fleet Vehicle
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

            <form onSubmit={handleCreateVehicle} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Registration # *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. DL-01-EV-1234"
                    value={formData.registration_number}
                    onChange={(e) => setFormData({ ...formData, registration_number: e.target.value })}
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
                    Vehicle Type *
                  </label>
                  <select
                    value={formData.vehicle_type}
                    onChange={(e) => setFormData({ ...formData, vehicle_type: e.target.value })}
                    style={{
                      width: '100%',
                      padding: '0.55rem 0.75rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(15, 23, 42, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                    }}
                  >
                    <option value="ELECTRIC_VAN">Electric Van (Eco-Friendly)</option>
                    <option value="VAN">Standard Cargo Van</option>
                    <option value="MOTORCYCLE">Two-Wheeler / Bike</option>
                    <option value="TRUCK">Heavy Logistics Truck</option>
                    <option value="THREE_WHEELER">Three-Wheeler Auto</option>
                  </select>
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Make & Model
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. Tata Ace EV / Mahindra Bolero"
                    value={formData.make_model}
                    onChange={(e) => setFormData({ ...formData, make_model: e.target.value })}
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
                    Model Year
                  </label>
                  <input
                    type="number"
                    min="1990"
                    max="2035"
                    value={formData.year}
                    onChange={(e) => setFormData({ ...formData, year: parseInt(e.target.value) || 2025 })}
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

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Payload Limit (kg) *
                  </label>
                  <input
                    type="number"
                    step="any"
                    min="1"
                    required
                    value={formData.max_weight_kg}
                    onChange={(e) => setFormData({ ...formData, max_weight_kg: parseFloat(e.target.value) || 500 })}
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
                    Volume Capacity (m³) *
                  </label>
                  <input
                    type="number"
                    step="any"
                    min="0.1"
                    required
                    value={formData.max_volume_cbm}
                    onChange={(e) => setFormData({ ...formData, max_volume_cbm: parseFloat(e.target.value) || 3.0 })}
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
                  Assigned Branch Hub
                </label>
                <select
                  value={formData.assigned_branch_id}
                  onChange={(e) => setFormData({ ...formData, assigned_branch_id: e.target.value })}
                  style={{
                    width: '100%',
                    padding: '0.55rem 0.75rem',
                    borderRadius: '6px',
                    border: '1px solid var(--border-subtle)',
                    background: 'rgba(15, 23, 42, 0.8)',
                    color: 'var(--text-primary)',
                    fontSize: '0.85rem',
                  }}
                >
                  <option value="">Central Depot (Floating)</option>
                  {branches.map((b) => (
                    <option key={b.id} value={b.id}>
                      {b.name} ({b.branch_code})
                    </option>
                  ))}
                </select>
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
                  {creating ? 'Saving...' : 'Register Vehicle'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Assign Driver Modal */}
      {showAssignModal && selectedVehicle && (
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
            maxWidth: '460px',
            padding: '1.75rem',
            boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                <UserCheck size={20} style={{ color: 'var(--accent-cyan)' }} />
                <h3 style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                  Assign Driver to {selectedVehicle.registration_number}
                </h3>
              </div>
              <button
                onClick={() => {
                  setShowAssignModal(false);
                  setSelectedVehicle(null);
                }}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '1.2rem' }}
              >
                ✕
              </button>
            </div>

            {assignError && (
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
                <span>{assignError}</span>
              </div>
            )}

            <form onSubmit={handleAssignDriver} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Select Qualified Driver *
                </label>
                <select
                  required
                  value={selectedDriverId}
                  onChange={(e) => setSelectedDriverId(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.75rem',
                    borderRadius: '6px',
                    border: '1px solid var(--border-subtle)',
                    background: 'rgba(15, 23, 42, 0.8)',
                    color: 'var(--text-primary)',
                    fontSize: '0.85rem',
                  }}
                >
                  <option value="">Choose an active driver...</option>
                  {drivers.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.first_name} {d.last_name} ({d.employee_code}){d.license_number ? ` • Lic: ${d.license_number}` : ''}
                    </option>
                  ))}
                </select>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1rem' }}>
                <button
                  type="button"
                  onClick={() => {
                    setShowAssignModal(false);
                    setSelectedVehicle(null);
                  }}
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
                  disabled={assigning || !selectedDriverId}
                  style={{
                    padding: '0.65rem 1.5rem',
                    borderRadius: '8px',
                    border: 'none',
                    background: 'linear-gradient(135deg, #0284c7 0%, #0369a1 100%)',
                    color: '#ffffff',
                    fontWeight: 600,
                    cursor: (assigning || !selectedDriverId) ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                  }}
                >
                  {assigning ? <Loader2 size={16} className="animate-spin" /> : null}
                  {assigning ? 'Assigning...' : 'Confirm Assignment'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

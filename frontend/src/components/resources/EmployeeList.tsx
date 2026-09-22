import React, { useState, useEffect, useCallback } from 'react';
import { 
  Users, 
  Plus, 
  Search, 
  AlertCircle, 
  CheckCircle2, 
  Loader2, 
  Building2, 
  Phone, 
  Mail, 
  ShieldCheck, 
  Trash2,
  SlidersHorizontal,
  FileBadge
} from 'lucide-react';
import type { Employee, CreateEmployeePayload, Branch } from '../../types/resources';
import { api } from '../../services/api';

interface EmployeeListProps {
  tenantId: string;
  userRole: string;
}

export const EmployeeList: React.FC<EmployeeListProps> = ({ tenantId, userRole }) => {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Search & Filters
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [branchFilter, setBranchFilter] = useState('');

  // Modal State
  const [showModal, setShowModal] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const [formData, setFormData] = useState<CreateEmployeePayload>({
    employee_code: '',
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    designation: '',
    employment_type: 'FULL_TIME',
    operational_role: 'DRIVER',
    license_number: '',
    branch_id: '',
  });

  const canManage = userRole === 'TENANT_ADMIN' || userRole === 'PLATFORM_ADMIN';

  // Fetch Branches for selection
  useEffect(() => {
    let isMounted = true;
    api.listBranches(tenantId, { limit: 100 })
      .then((res) => {
        if (isMounted) {
          setBranches(res.branches || []);
        }
      })
      .catch(() => {});
    return () => {
      isMounted = false;
    };
  }, [tenantId]);

  const fetchEmployees = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listEmployees(tenantId, {
        search: searchTerm || undefined,
        operational_role: roleFilter || undefined,
        branch_id: branchFilter || undefined,
        status: statusFilter || undefined,
        limit: 50,
      });
      setEmployees(res.employees || []);
      setTotal(res.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to load employee profiles');
    } finally {
      setLoading(false);
    }
  }, [tenantId, searchTerm, roleFilter, branchFilter, statusFilter]);

  useEffect(() => {
    fetchEmployees();
  }, [fetchEmployees]);

  const handleCreateEmployee = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setCreating(true);

    try {
      const newEmp = await api.createEmployee(tenantId, {
        ...formData,
        employee_code: formData.employee_code.trim().toUpperCase(),
        email: formData.email?.trim() || undefined,
        phone: formData.phone?.trim() || undefined,
        license_number: formData.license_number?.trim() || undefined,
        branch_id: formData.branch_id || undefined,
      });
      setEmployees((prev) => [newEmp, ...prev]);
      setTotal((prev) => prev + 1);
      setSuccess(`Employee "${newEmp.first_name} ${newEmp.last_name}" (${newEmp.employee_code}) registered successfully`);
      setShowModal(false);
      setFormData({
        employee_code: '',
        first_name: '',
        last_name: '',
        email: '',
        phone: '',
        designation: '',
        employment_type: 'FULL_TIME',
        operational_role: 'DRIVER',
        license_number: '',
        branch_id: '',
      });
    } catch (err: any) {
      setFormError(err.message || 'Failed to register employee');
    } finally {
      setCreating(false);
    }
  };

  const handleDeactivate = async (empId: string, name: string) => {
    if (!confirm(`Are you sure you want to deactivate employee "${name}"?`)) {
      return;
    }

    try {
      await api.deleteEmployee(tenantId, empId);
      setEmployees((prev) =>
        prev.map((emp) => (emp.id === empId ? { ...emp, is_active: false, status: 'TERMINATED' } : emp))
      );
      setSuccess(`Employee "${name}" deactivated successfully`);
    } catch (err: any) {
      setError(err.message || 'Failed to deactivate employee');
    }
  };

  const getRoleBadgeStyle = (role: string) => {
    switch (role) {
      case 'DRIVER':
        return { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)' };
      case 'OPERATOR':
        return { bg: 'rgba(16, 185, 129, 0.15)', text: '#34d399', border: 'rgba(16, 185, 129, 0.3)' };
      case 'DISPATCHER':
        return { bg: 'rgba(245, 158, 11, 0.15)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.3)' };
      case 'SUPERVISOR':
      case 'MANAGER':
        return { bg: 'rgba(168, 85, 247, 0.15)', text: '#c084fc', border: 'rgba(168, 85, 247, 0.3)' };
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
            <Users size={22} />
          </div>
          <div>
            <h2 style={{ fontSize: '1.35rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
              Staff & Driver Workforce
            </h2>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', margin: 0 }}>
              {total} active logistics personnel assigned to delivery hubs
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
              background: 'linear-gradient(135deg, #06b6d4 0%, #0284c7 100%)',
              color: '#ffffff',
              fontWeight: 600,
              fontSize: '0.9rem',
              cursor: 'pointer',
              boxShadow: '0 4px 14px rgba(6, 182, 212, 0.35)',
              transition: 'all 0.2s ease',
            }}
          >
            <Plus size={18} />
            Onboard Employee
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
            placeholder="Search by code, name, designation..."
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
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.85rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Roles</option>
            <option value="DRIVER">Drivers</option>
            <option value="OPERATOR">Operators</option>
            <option value="DISPATCHER">Dispatchers</option>
            <option value="SUPERVISOR">Supervisors</option>
            <option value="MANAGER">Managers</option>
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
            <option value="ACTIVE">Active</option>
            <option value="ON_LEAVE">On Leave</option>
            <option value="SUSPENDED">Suspended</option>
            <option value="TERMINATED">Terminated</option>
          </select>
        </div>
      </div>

      {/* Employees Grid */}
      {loading ? (
        <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          <Loader2 size={32} className="animate-spin" style={{ margin: '0 auto 1rem', display: 'block' }} />
          <p>Loading staff and drivers...</p>
        </div>
      ) : employees.length === 0 ? (
        <div style={{
          padding: '4rem 2rem',
          textAlign: 'center',
          background: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px dashed var(--border-subtle)',
        }}>
          <Users size={48} style={{ color: 'var(--text-muted)', margin: '0 auto 1rem', opacity: 0.5 }} />
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
            No personnel found
          </h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', maxWidth: '400px', margin: '0 auto 1.5rem' }}>
            {searchTerm ? 'No employees match your search criteria.' : 'Start assembling your logistics team by registering drivers and operators.'}
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
              Onboard First Employee
            </button>
          )}
        </div>
      ) : (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))',
          gap: '1.25rem',
        }}>
          {employees.map((emp) => {
            const roleStyle = getRoleBadgeStyle(emp.operational_role);
            return (
              <div
                key={emp.id}
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
                      background: 'rgba(6, 182, 212, 0.15)',
                      color: 'var(--accent-cyan)',
                      border: '1px solid rgba(6, 182, 212, 0.3)',
                      display: 'inline-block',
                      marginBottom: '0.4rem',
                    }}>
                      {emp.employee_code}
                    </span>
                    <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
                      {emp.first_name} {emp.last_name}
                    </h3>
                    <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                      {emp.designation}
                    </span>
                  </div>

                  <span style={{
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    padding: '0.25rem 0.6rem',
                    borderRadius: '12px',
                    background: roleStyle.bg,
                    color: roleStyle.text,
                    border: `1px solid ${roleStyle.border}`,
                  }}>
                    {emp.operational_role}
                  </span>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.45rem', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                  {emp.branch_name && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <Building2 size={16} style={{ color: 'var(--accent-blue)', flexShrink: 0 }} />
                      <span>{emp.branch_name} {emp.branch_code ? `(${emp.branch_code})` : ''}</span>
                    </div>
                  )}

                  {emp.email && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <Mail size={16} style={{ color: 'var(--accent-cyan)', flexShrink: 0 }} />
                      <span>{emp.email}</span>
                    </div>
                  )}

                  {emp.phone && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <Phone size={16} style={{ color: 'var(--accent-emerald)', flexShrink: 0 }} />
                      <span>{emp.phone}</span>
                    </div>
                  )}

                  {emp.license_number && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <FileBadge size={16} style={{ color: 'var(--accent-amber)', flexShrink: 0 }} />
                      <span>Lic: <strong>{emp.license_number}</strong></span>
                    </div>
                  )}

                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.25rem' }}>
                    <ShieldCheck size={16} style={{ color: emp.is_active ? 'var(--accent-emerald)' : 'var(--accent-rose)', flexShrink: 0 }} />
                    <span style={{ fontSize: '0.75rem', fontWeight: 600, color: emp.is_active ? '#34d399' : '#fb7185' }}>
                      {emp.status} ({emp.employment_type})
                    </span>
                  </div>
                </div>

                {canManage && emp.is_active && (
                  <div style={{ marginTop: 'auto', paddingTop: '0.75rem', borderTop: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'flex-end' }}>
                    <button
                      onClick={() => handleDeactivate(emp.id, `${emp.first_name} ${emp.last_name}`)}
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
                      Deactivate
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Add Employee Modal */}
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
                <Users size={22} style={{ color: 'var(--accent-cyan)' }} />
                <h3 style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                  Onboard Logistics Personnel
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

            <form onSubmit={handleCreateEmployee} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Staff Code *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. DRV-101"
                    value={formData.employee_code}
                    onChange={(e) => setFormData({ ...formData, employee_code: e.target.value })}
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
                    Designation *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Heavy Duty Express Driver"
                    value={formData.designation}
                    onChange={(e) => setFormData({ ...formData, designation: e.target.value })}
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
                    First Name *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="First name"
                    value={formData.first_name}
                    onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
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
                    Last Name *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Last name"
                    value={formData.last_name}
                    onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
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
                    Email
                  </label>
                  <input
                    type="email"
                    placeholder="driver@company.com"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
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
                    Phone
                  </label>
                  <input
                    type="tel"
                    placeholder="+91 98765 43210"
                    value={formData.phone}
                    onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
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
                    Operational Role *
                  </label>
                  <select
                    value={formData.operational_role}
                    onChange={(e) => setFormData({ ...formData, operational_role: e.target.value })}
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
                    <option value="DRIVER">Fleet Driver</option>
                    <option value="OPERATOR">Hub Operator</option>
                    <option value="DISPATCHER">Dispatcher</option>
                    <option value="SUPERVISOR">Supervisor</option>
                    <option value="MANAGER">Manager</option>
                  </select>
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Employment Type *
                  </label>
                  <select
                    value={formData.employment_type}
                    onChange={(e) => setFormData({ ...formData, employment_type: e.target.value })}
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
                    <option value="FULL_TIME">Full Time</option>
                    <option value="PART_TIME">Part Time</option>
                    <option value="CONTRACTOR">Contractor</option>
                    <option value="INTERN">Intern</option>
                  </select>
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    Commercial Driver License #
                  </label>
                  <input
                    type="text"
                    placeholder="DL-04-2022-0091823"
                    value={formData.license_number}
                    onChange={(e) => setFormData({ ...formData, license_number: e.target.value })}
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
                    Assigned Delivery Hub
                  </label>
                  <select
                    value={formData.branch_id}
                    onChange={(e) => setFormData({ ...formData, branch_id: e.target.value })}
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
                    <option value="">No Hub Assignment (Float)</option>
                    {branches.map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.name} ({b.branch_code})
                      </option>
                    ))}
                  </select>
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
                    background: 'linear-gradient(135deg, #06b6d4 0%, #0284c7 100%)',
                    color: '#ffffff',
                    fontWeight: 600,
                    cursor: creating ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                  }}
                >
                  {creating ? <Loader2 size={16} className="animate-spin" /> : null}
                  {creating ? 'Saving...' : 'Register Employee'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

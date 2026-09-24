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
  FileBadge,
  Edit3,
  Activity,
  Key,
  UserCheck
} from 'lucide-react';
import type { Employee, CreateEmployeePayload, Branch, EmployeeAccountStatus } from '../../types/resources';
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
  const [availabilityFilter, setAvailabilityFilter] = useState('');
  const [availableDriversOnly, setAvailableDriversOnly] = useState(false);

  // Modal State
  const [showModal, setShowModal] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  // Status Update Modal State
  const [statusModalEmp, setStatusModalEmp] = useState<Employee | null>(null);
  const [statusFormData, setStatusFormData] = useState<{
    status: Employee['status'];
    availability_status: Employee['availability_status'];
    verification_status: Employee['verification_status'];
  }>({
    status: 'ACTIVE',
    availability_status: 'AVAILABLE',
    verification_status: 'VERIFIED',
  });
  const [updatingStatus, setUpdatingStatus] = useState(false);
  const [statusUpdateError, setStatusUpdateError] = useState<string | null>(null);

  const [formData, setFormData] = useState<CreateEmployeePayload>({
    employee_code: '',
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    designation: '',
    employment_type: 'FULL_TIME',
    operational_role: 'DRIVER',
    availability_status: 'AVAILABLE',
    verification_status: 'VERIFIED',
    license_number: '',
    branch_id: '',
  });

  // Account Provisioning State
  const [createWithAccount, setCreateWithAccount] = useState(false);
  const [accountPassword, setAccountPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [accountSystemRole, setAccountSystemRole] = useState('EMPLOYEE');
  const [sendInvite, setSendInvite] = useState(false);

  // Account Status Inspection Modal
  const [accountStatusModalEmp, setAccountStatusModalEmp] = useState<Employee | null>(null);
  const [accountStatusData, setAccountStatusData] = useState<EmployeeAccountStatus | null>(null);
  const [accountStatusLoading, setAccountStatusLoading] = useState(false);
  const [accountStatusError, setAccountStatusError] = useState<string | null>(null);

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
      if (availableDriversOnly) {
        const res = await api.listAvailableDrivers(tenantId, branchFilter || undefined);
        setEmployees(res.drivers || []);
        setTotal(res.total || 0);
      } else {
        const res = await api.listEmployees(tenantId, {
          search: searchTerm || undefined,
          operational_role: roleFilter || undefined,
          branch_id: branchFilter || undefined,
          status: statusFilter || undefined,
          limit: 50,
        });
        let list = res.employees || [];
        if (availabilityFilter) {
          list = list.filter((e) => e.availability_status === availabilityFilter);
        }
        setEmployees(list);
        setTotal(list.length);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load employee profiles');
    } finally {
      setLoading(false);
    }
  }, [tenantId, searchTerm, roleFilter, branchFilter, statusFilter, availabilityFilter, availableDriversOnly]);

  useEffect(() => {
    fetchEmployees();
  }, [fetchEmployees]);

  const handleCreateEmployee = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setCreating(true);

    try {
      const codeTrimmed = formData.employee_code?.trim();
      let newEmp: Employee;

      if (createWithAccount) {
        if (!formData.email?.trim()) {
          setFormError('An email address is required to provision a login account.');
          setCreating(false);
          return;
        }
        if (!sendInvite && (!accountPassword || accountPassword.length < 8)) {
          setFormError('Temporary password must be at least 8 characters long (or enable Invitation link).');
          setCreating(false);
          return;
        }
        newEmp = await api.createEmployeeWithAccount(tenantId, {
          ...formData,
          employee_code: codeTrimmed ? codeTrimmed.toUpperCase() : undefined,
          email: formData.email.trim(),
          phone: formData.phone?.trim() || undefined,
          license_number: formData.license_number?.trim() || undefined,
          branch_id: formData.branch_id || undefined,
          password: accountPassword || undefined,
          system_role: accountSystemRole,
          send_invite: sendInvite,
        });
      } else {
        newEmp = await api.createEmployee(tenantId, {
          ...formData,
          employee_code: codeTrimmed ? codeTrimmed.toUpperCase() : undefined,
          email: formData.email?.trim() || undefined,
          phone: formData.phone?.trim() || undefined,
          license_number: formData.license_number?.trim() || undefined,
          branch_id: formData.branch_id || undefined,
        });
      }

      setEmployees((prev) => [newEmp, ...prev]);
      setTotal((prev) => prev + 1);
      setSuccess(`Employee "${newEmp.first_name} ${newEmp.last_name}" (${newEmp.employee_code}) registered successfully${createWithAccount ? ' with login credentials' : ''}`);
      setShowModal(false);
      setCreateWithAccount(false);
      setAccountPassword('');
      setAccountSystemRole('EMPLOYEE');
      setSendInvite(false);
      setFormData({
        employee_code: '',
        first_name: '',
        last_name: '',
        email: '',
        phone: '',
        designation: '',
        employment_type: 'FULL_TIME',
        operational_role: 'DRIVER',
        availability_status: 'AVAILABLE',
        verification_status: 'VERIFIED',
        license_number: '',
        branch_id: '',
      });
    } catch (err: any) {
      setFormError(err.message || 'Failed to register employee');
    } finally {
      setCreating(false);
    }
  };

  const handleInspectAccount = async (emp: Employee) => {
    setAccountStatusModalEmp(emp);
    setAccountStatusLoading(true);
    setAccountStatusError(null);
    setAccountStatusData(null);
    try {
      const status = await api.getEmployeeAccountStatus(tenantId, emp.id);
      setAccountStatusData(status);
    } catch (err: any) {
      setAccountStatusError(err.message || 'Failed to fetch account status');
    } finally {
      setAccountStatusLoading(false);
    }
  };

  const handleOpenStatusModal = (emp: Employee) => {
    setStatusModalEmp(emp);
    setStatusFormData({
      status: emp.status || 'ACTIVE',
      availability_status: emp.availability_status || 'AVAILABLE',
      verification_status: emp.verification_status || 'VERIFIED',
    });
    setStatusUpdateError(null);
  };

  const handleSaveStatus = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!statusModalEmp) return;
    setUpdatingStatus(true);
    setStatusUpdateError(null);
    try {
      const updated = await api.updateEmployeeStatus(tenantId, statusModalEmp.id, statusFormData);
      setEmployees((prev) =>
        prev.map((emp) => (emp.id === statusModalEmp.id ? { ...emp, ...updated } : emp))
      );
      setSuccess(`Updated status for employee ${statusModalEmp.employee_code}`);
      setStatusModalEmp(null);
    } catch (err: any) {
      setStatusUpdateError(err.message || 'Failed to update employee status');
    } finally {
      setUpdatingStatus(false);
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
      case 'DELIVERY_EXECUTIVE':
        return { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)' };
      case 'OPERATOR':
      case 'WAREHOUSE_OPERATOR':
        return { bg: 'rgba(16, 185, 129, 0.15)', text: '#34d399', border: 'rgba(16, 185, 129, 0.3)' };
      case 'DISPATCHER':
        return { bg: 'rgba(245, 158, 11, 0.15)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.3)' };
      case 'SUPERVISOR':
      case 'MANAGER':
      case 'BRANCH_MANAGER':
        return { bg: 'rgba(168, 85, 247, 0.15)', text: '#c084fc', border: 'rgba(168, 85, 247, 0.3)' };
      default:
        return { bg: 'rgba(148, 163, 184, 0.15)', text: '#94a3b8', border: 'rgba(148, 163, 184, 0.3)' };
    }
  };

  const getAvailabilityBadgeStyle = (avail?: string) => {
    switch (avail) {
      case 'AVAILABLE':
        return { bg: 'rgba(16, 185, 129, 0.15)', text: '#34d399', border: 'rgba(16, 185, 129, 0.3)', dot: '#10b981', label: 'Available' };
      case 'BUSY':
        return { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)', dot: '#3b82f6', label: 'On Route / Busy' };
      case 'OFF_DUTY':
        return { bg: 'rgba(245, 158, 11, 0.15)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.3)', dot: '#f59e0b', label: 'Off Duty' };
      case 'UNAVAILABLE':
      default:
        return { bg: 'rgba(244, 63, 94, 0.15)', text: '#fda4af', border: 'rgba(244, 63, 94, 0.3)', dot: '#f43f5e', label: 'Unavailable' };
    }
  };

  const getVerificationBadgeStyle = (ver?: string) => {
    switch (ver) {
      case 'VERIFIED':
        return { bg: 'rgba(16, 185, 129, 0.1)', text: '#34d399', border: 'rgba(16, 185, 129, 0.25)', label: 'KYC Verified' };
      case 'REJECTED':
        return { bg: 'rgba(244, 63, 94, 0.1)', text: '#fda4af', border: 'rgba(244, 63, 94, 0.25)', label: 'KYC Rejected' };
      case 'PENDING':
      default:
        return { bg: 'rgba(245, 158, 11, 0.1)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.25)', label: 'KYC Pending' };
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
            <option value="BRANCH_MANAGER">Branch Managers</option>
            <option value="WAREHOUSE_OPERATOR">Warehouse Operators</option>
            <option value="DELIVERY_EXECUTIVE">Delivery Executives</option>
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

          <select
            value={availabilityFilter}
            onChange={(e) => {
              setAvailabilityFilter(e.target.value);
              setAvailableDriversOnly(false);
            }}
            style={{
              padding: '0.55rem 0.85rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Availability</option>
            <option value="AVAILABLE">Available</option>
            <option value="BUSY">Busy</option>
            <option value="OFF_DUTY">Off Duty</option>
            <option value="UNAVAILABLE">Unavailable</option>
          </select>

          <button
            type="button"
            onClick={() => setAvailableDriversOnly(!availableDriversOnly)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              padding: '0.55rem 0.85rem',
              borderRadius: '6px',
              border: availableDriversOnly ? '1px solid rgba(16, 185, 129, 0.4)' : '1px solid var(--border-subtle)',
              background: availableDriversOnly ? 'rgba(16, 185, 129, 0.15)' : 'rgba(15, 23, 42, 0.6)',
              color: availableDriversOnly ? '#34d399' : 'var(--text-secondary)',
              fontSize: '0.85rem',
              fontWeight: 600,
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
          >
            <Activity size={14} />
            Available Drivers Only
          </button>
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
            const availStyle = getAvailabilityBadgeStyle(emp.availability_status);
            const verStyle = getVerificationBadgeStyle(emp.verification_status);

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
                    <div style={{ marginTop: '0.35rem' }}>
                      {emp.user_id ? (
                        <button
                          onClick={() => handleInspectAccount(emp)}
                          title="Click to inspect linked login credentials and role status"
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '0.3rem',
                            padding: '0.15rem 0.5rem',
                            borderRadius: '6px',
                            background: 'rgba(16, 185, 129, 0.12)',
                            color: '#34d399',
                            border: '1px solid rgba(16, 185, 129, 0.3)',
                            fontSize: '0.7rem',
                            fontWeight: 600,
                            cursor: 'pointer',
                          }}
                        >
                          <Key size={11} />
                          <span>Login Account Active</span>
                        </button>
                      ) : (
                        <span style={{
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: '0.3rem',
                          padding: '0.15rem 0.5rem',
                          borderRadius: '6px',
                          background: 'rgba(148, 163, 184, 0.08)',
                          color: '#94a3b8',
                          border: '1px solid rgba(148, 163, 184, 0.2)',
                          fontSize: '0.7rem',
                          fontWeight: 500,
                        }}>
                          Profile Only
                        </span>
                      )}
                    </div>
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.35rem' }}>
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

                    <span style={{
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      padding: '0.2rem 0.5rem',
                      borderRadius: '12px',
                      background: availStyle.bg,
                      color: availStyle.text,
                      border: `1px solid ${availStyle.border}`,
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '0.35rem',
                    }}>
                      <span style={{ width: '6px', height: '6px', borderRadius: '50%', background: availStyle.dot }} />
                      {availStyle.label}
                    </span>
                  </div>
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

                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '0.5rem', marginTop: '0.25rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <ShieldCheck size={16} style={{ color: emp.is_active ? 'var(--accent-emerald)' : 'var(--accent-rose)', flexShrink: 0 }} />
                      <span style={{ fontSize: '0.75rem', fontWeight: 600, color: emp.is_active ? '#34d399' : '#fb7185' }}>
                        {emp.status} ({emp.employment_type})
                      </span>
                    </div>

                    <span style={{
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      padding: '0.15rem 0.45rem',
                      borderRadius: '8px',
                      background: verStyle.bg,
                      color: verStyle.text,
                      border: `1px solid ${verStyle.border}`,
                    }}>
                      {verStyle.label}
                    </span>
                  </div>
                </div>

                <div style={{ marginTop: 'auto', paddingTop: '0.75rem', borderTop: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  {canManage ? (
                    <button
                      onClick={() => handleOpenStatusModal(emp)}
                      style={{
                        background: 'rgba(6, 182, 212, 0.1)',
                        border: '1px solid rgba(6, 182, 212, 0.25)',
                        color: 'var(--accent-cyan)',
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.35rem',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        cursor: 'pointer',
                        padding: '0.35rem 0.65rem',
                        borderRadius: '6px',
                        transition: 'all 0.2s ease',
                      }}
                      onMouseEnter={(e) => (e.currentTarget.style.background = 'rgba(6, 182, 212, 0.2)')}
                      onMouseLeave={(e) => (e.currentTarget.style.background = 'rgba(6, 182, 212, 0.1)')}
                    >
                      <Edit3 size={12} />
                      Update Status
                    </button>
                  ) : <div />}

                  {canManage && emp.is_active && (
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
                  )}
                </div>
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
                    Staff Code <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>(Auto if blank)</span>
                  </label>
                  <input
                    type="text"
                    placeholder="Leave empty for EMP-XXXX"
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
                    <option value="BRANCH_MANAGER">Branch Manager</option>
                    <option value="WAREHOUSE_OPERATOR">Warehouse Operator</option>
                    <option value="DELIVERY_EXECUTIVE">Delivery Executive</option>
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
                    Operational Availability *
                  </label>
                  <select
                    value={formData.availability_status}
                    onChange={(e) => setFormData({ ...formData, availability_status: e.target.value })}
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
                    <option value="AVAILABLE">Available</option>
                    <option value="BUSY">Busy</option>
                    <option value="OFF_DUTY">Off Duty</option>
                    <option value="UNAVAILABLE">Unavailable</option>
                  </select>
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                    KYC Verification Status *
                  </label>
                  <select
                    value={formData.verification_status}
                    onChange={(e) => setFormData({ ...formData, verification_status: e.target.value })}
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
                    <option value="VERIFIED">KYC Verified</option>
                    <option value="PENDING">KYC Pending</option>
                    <option value="REJECTED">KYC Rejected</option>
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

              {/* Provision Login Account Section */}
              <div style={{
                marginTop: '0.5rem',
                padding: '1rem',
                borderRadius: '8px',
                background: 'rgba(6, 182, 212, 0.05)',
                border: '1px solid rgba(6, 182, 212, 0.2)',
                display: 'flex',
                flexDirection: 'column',
                gap: '0.75rem',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer', margin: 0 }}>
                    <input
                      type="checkbox"
                      checked={createWithAccount}
                      onChange={(e) => setCreateWithAccount(e.target.checked)}
                      style={{ cursor: 'pointer', accentColor: 'var(--accent-cyan)' }}
                    />
                    <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-primary)' }}>
                      Provision System Login Account
                    </span>
                  </label>
                  <span style={{ fontSize: '0.75rem', color: 'var(--accent-cyan)', fontWeight: 500 }}>
                    {createWithAccount ? 'Transactional Auth Linking Enabled' : 'Profile Only'}
                  </span>
                </div>

                {createWithAccount && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', marginTop: '0.5rem' }}>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                      Creates user credentials and tenant membership atomically. Passwords are encrypted with Bcrypt (cost 12). No password hashes or plain-text credentials are ever returned or logged.
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                      <div>
                        <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                          System Authorization Role *
                        </label>
                        <select
                          value={accountSystemRole}
                          onChange={(e) => setAccountSystemRole(e.target.value)}
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
                          <option value="EMPLOYEE">EMPLOYEE (Standard Staff / Driver)</option>
                          <option value="TENANT_OPERATOR">TENANT_OPERATOR (Hub Dispatcher)</option>
                          <option value="VIEWER">VIEWER (Read-Only Access)</option>
                          <option value="TENANT_ADMIN">TENANT_ADMIN (Full Tenant Admin)</option>
                        </select>
                      </div>

                      <div>
                        <label style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                          <span>Temporary Password *</span>
                          <button
                            type="button"
                            onClick={() => setShowPassword(!showPassword)}
                            style={{ background: 'none', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontSize: '0.75rem' }}
                          >
                            {showPassword ? 'Hide' : 'Show'}
                          </button>
                        </label>
                        <input
                          type={showPassword ? 'text' : 'password'}
                          required={!sendInvite}
                          disabled={sendInvite}
                          placeholder={sendInvite ? 'Auto-generated on invite' : 'Min 8 characters'}
                          value={accountPassword}
                          onChange={(e) => setAccountPassword(e.target.value)}
                          style={{
                            width: '100%',
                            padding: '0.55rem 0.75rem',
                            borderRadius: '6px',
                            border: '1px solid var(--border-subtle)',
                            background: sendInvite ? 'rgba(15, 23, 42, 0.4)' : 'rgba(15, 23, 42, 0.8)',
                            color: 'var(--text-primary)',
                            fontSize: '0.85rem',
                          }}
                        />
                      </div>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.25rem' }}>
                      <input
                        type="checkbox"
                        id="sendInvite"
                        checked={sendInvite}
                        onChange={(e) => setSendInvite(e.target.checked)}
                        style={{ cursor: 'pointer', accentColor: 'var(--accent-cyan)' }}
                      />
                      <label htmlFor="sendInvite" style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', cursor: 'pointer', margin: 0 }}>
                        Generate secure onboarding invitation token and email link
                      </label>
                    </div>
                  </div>
                )}
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

      {/* Update Employee Status Modal */}
      {statusModalEmp && (
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
            maxWidth: '480px',
            padding: '1.75rem',
            boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                <Edit3 size={20} style={{ color: 'var(--accent-cyan)' }} />
                <div>
                  <h3 style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                    Update Status & Availability
                  </h3>
                  <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', margin: 0 }}>
                    {statusModalEmp.first_name} {statusModalEmp.last_name} ({statusModalEmp.employee_code})
                  </p>
                </div>
              </div>
              <button
                onClick={() => setStatusModalEmp(null)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '1.2rem' }}
              >
                ✕
              </button>
            </div>

            {statusUpdateError && (
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
                <span>{statusUpdateError}</span>
              </div>
            )}

            <form onSubmit={handleSaveStatus} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Employment Status
                </label>
                <select
                  value={statusFormData.status}
                  onChange={(e) => setStatusFormData({ ...statusFormData, status: e.target.value as any })}
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
                  <option value="ACTIVE">ACTIVE</option>
                  <option value="ON_LEAVE">ON_LEAVE</option>
                  <option value="SUSPENDED">SUSPENDED</option>
                  <option value="TERMINATED">TERMINATED</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  Operational Availability
                </label>
                <select
                  value={statusFormData.availability_status}
                  onChange={(e) => setStatusFormData({ ...statusFormData, availability_status: e.target.value as any })}
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
                  <option value="AVAILABLE">AVAILABLE (Eligible for dispatch/vehicle assignment)</option>
                  <option value="BUSY">BUSY (Active in transit or assigned)</option>
                  <option value="OFF_DUTY">OFF_DUTY (Off shift)</option>
                  <option value="UNAVAILABLE">UNAVAILABLE (Leave or medical)</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', fontSize: '0.8rem', color: 'var(--text-secondary)', marginBottom: '0.35rem' }}>
                  KYC Verification Status
                </label>
                <select
                  value={statusFormData.verification_status}
                  onChange={(e) => setStatusFormData({ ...statusFormData, verification_status: e.target.value as any })}
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
                  <option value="VERIFIED">VERIFIED</option>
                  <option value="PENDING">PENDING</option>
                  <option value="REJECTED">REJECTED</option>
                </select>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1rem' }}>
                <button
                  type="button"
                  onClick={() => setStatusModalEmp(null)}
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
                  disabled={updatingStatus}
                  style={{
                    padding: '0.65rem 1.5rem',
                    borderRadius: '8px',
                    border: 'none',
                    background: 'linear-gradient(135deg, #06b6d4 0%, #0284c7 100%)',
                    color: '#ffffff',
                    fontWeight: 600,
                    cursor: updatingStatus ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                  }}
                >
                  {updatingStatus ? <Loader2 size={16} className="animate-spin" /> : null}
                  {updatingStatus ? 'Updating...' : 'Save Status'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Account Status Inspection Modal */}
      {accountStatusModalEmp && (
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
            maxWidth: '480px',
            padding: '1.75rem',
            boxShadow: '0 20px 50px rgba(0,0,0,0.5)',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
                <Key size={20} style={{ color: 'var(--accent-cyan)' }} />
                <div>
                  <h3 style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                    Employee Account Status
                  </h3>
                  <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', margin: 0 }}>
                    {accountStatusModalEmp.first_name} {accountStatusModalEmp.last_name} ({accountStatusModalEmp.employee_code})
                  </p>
                </div>
              </div>
              <button
                onClick={() => setAccountStatusModalEmp(null)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '1.2rem' }}
              >
                ✕
              </button>
            </div>

            {accountStatusLoading ? (
              <div style={{ textAlign: 'center', padding: '2rem' }}>
                <Loader2 size={24} className="animate-spin" style={{ margin: '0 auto 0.5rem', color: 'var(--accent-cyan)' }} />
                <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>Fetching account link status...</p>
              </div>
            ) : accountStatusError ? (
              <div style={{
                padding: '0.75rem 1rem',
                borderRadius: '8px',
                background: 'rgba(244, 63, 94, 0.1)',
                border: '1px solid rgba(244, 63, 94, 0.3)',
                color: '#fda4af',
                fontSize: '0.85rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
              }}>
                <AlertCircle size={16} />
                <span>{accountStatusError}</span>
              </div>
            ) : accountStatusData ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem', fontSize: '0.85rem' }}>
                <div style={{
                  padding: '0.75rem',
                  borderRadius: '8px',
                  background: accountStatusData.has_account ? 'rgba(16, 185, 129, 0.1)' : 'rgba(245, 158, 11, 0.1)',
                  border: `1px solid ${accountStatusData.has_account ? 'rgba(16, 185, 129, 0.25)' : 'rgba(245, 158, 11, 0.25)'}`,
                  color: accountStatusData.has_account ? '#34d399' : '#fbbf24',
                  fontWeight: 600,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                }}>
                  <UserCheck size={16} />
                  <span>
                    {accountStatusData.has_account ? 'Login Credentials Linked & Configured' : 'No User Account Linked (Profile Only)'}
                  </span>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
                  <div style={{ padding: '0.5rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.6)' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block' }}>Operational Role</span>
                    <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{accountStatusData.operational_role}</span>
                  </div>
                  <div style={{ padding: '0.5rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.6)' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block' }}>System Role</span>
                    <span style={{ fontWeight: 600, color: 'var(--accent-cyan)' }}>{accountStatusData.system_role || 'None'}</span>
                  </div>
                </div>

                {accountStatusData.user_email && (
                  <div style={{ padding: '0.5rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.6)' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block' }}>Login Email</span>
                    <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{accountStatusData.user_email}</span>
                  </div>
                )}

                {accountStatusData.branch_name && (
                  <div style={{ padding: '0.5rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.6)' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'block' }}>Branch Hub</span>
                    <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{accountStatusData.branch_name}</span>
                  </div>
                )}

                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '0.5rem' }}>
                  <button
                    onClick={() => setAccountStatusModalEmp(null)}
                    style={{
                      padding: '0.55rem 1.25rem',
                      borderRadius: '8px',
                      border: '1px solid var(--border-subtle)',
                      background: 'transparent',
                      color: 'var(--text-secondary)',
                      fontWeight: 600,
                      cursor: 'pointer',
                    }}
                  >
                    Close
                  </button>
                </div>
              </div>
            ) : null}
          </div>
        </div>
      )}
    </div>
  );
};

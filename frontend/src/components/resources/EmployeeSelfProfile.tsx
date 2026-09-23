import React, { useState, useEffect } from 'react';
import {
  User,
  Truck,
  Building2,
  ShieldCheck,
  Mail,
  Phone,
  FileBadge,
  Calendar,
  Activity,
  AlertCircle,
  Loader2
} from 'lucide-react';
import type { EmployeeMeResponse } from '../../types/resources';
import { api } from '../../services/api';

interface EmployeeSelfProfileProps {
  tenantId: string;
}

export const EmployeeSelfProfile: React.FC<EmployeeSelfProfileProps> = ({ tenantId }) => {
  const [profile, setProfile] = useState<EmployeeMeResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;
    setLoading(true);
    setError(null);

    api.getMyProfile(tenantId)
      .then((res) => {
        if (isMounted) {
          setProfile(res);
        }
      })
      .catch((err) => {
        if (isMounted) {
          setError(err.message || 'No linked operational employee profile found for this account.');
        }
      })
      .finally(() => {
        if (isMounted) {
          setLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, [tenantId]);

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '300px' }}>
        <Loader2 size={32} className="spin-slow" style={{ color: 'var(--accent-cyan)' }} />
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div style={{
        maxWidth: '650px',
        margin: '2rem auto',
        padding: '2rem',
        borderRadius: '12px',
        background: 'var(--bg-card)',
        border: '1px solid var(--border-subtle)',
        textAlign: 'center',
      }}>
        <AlertCircle size={40} style={{ color: 'var(--accent-amber)', margin: '0 auto 1rem' }} />
        <h3 style={{ fontSize: '1.2rem', color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
          Operational Profile Notice
        </h3>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '1.5rem' }}>
          {error || 'No employee operational profile is currently linked to your user account in this organization.'}
        </p>
      </div>
    );
  }

  const { employee, system_role, assigned_vehicle } = profile;

  return (
    <div style={{ maxWidth: '850px', margin: '0 auto', display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Header Profile Card */}
      <div style={{
        background: 'var(--bg-card)',
        borderRadius: '16px',
        border: '1px solid var(--border-subtle)',
        padding: '2rem',
        display: 'flex',
        flexWrap: 'wrap',
        justifyContent: 'space-between',
        alignItems: 'center',
        gap: '1.5rem',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
          <div style={{
            width: '64px',
            height: '64px',
            borderRadius: '16px',
            background: 'linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(59, 130, 246, 0.2))',
            border: '1px solid rgba(6, 182, 212, 0.4)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--accent-cyan)',
          }}>
            <User size={32} />
          </div>

          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.25rem' }}>
              <h2 style={{ fontSize: '1.4rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                {employee.first_name} {employee.last_name}
              </h2>
              <span style={{
                fontSize: '0.75rem',
                fontWeight: 700,
                padding: '0.2rem 0.5rem',
                borderRadius: '6px',
                background: 'rgba(6, 182, 212, 0.15)',
                color: 'var(--accent-cyan)',
                border: '1px solid rgba(6, 182, 212, 0.3)',
              }}>
                {employee.employee_code}
              </span>
            </div>
            <p style={{ margin: 0, fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
              {employee.designation} &bull; <strong style={{ color: 'var(--text-primary)' }}>{employee.operational_role}</strong>
            </p>
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.5rem' }}>
          <span style={{
            fontSize: '0.75rem',
            fontWeight: 700,
            padding: '0.35rem 0.85rem',
            borderRadius: '20px',
            background: employee.availability_status === 'AVAILABLE' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(245, 158, 11, 0.15)',
            color: employee.availability_status === 'AVAILABLE' ? '#34d399' : '#fbbf24',
            border: `1px solid ${employee.availability_status === 'AVAILABLE' ? 'rgba(16, 185, 129, 0.3)' : 'rgba(245, 158, 11, 0.3)'}`,
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.4rem',
          }}>
            <Activity size={14} />
            {employee.availability_status}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            System Role: <strong>{system_role}</strong>
          </span>
        </div>
      </div>

      {/* Grid: Branch & Assigned Fleet Vehicle */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(360px, 1fr))', gap: '1.5rem' }}>
        {/* Branch Card */}
        <div style={{
          background: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px solid var(--border-subtle)',
          padding: '1.5rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <Building2 size={22} style={{ color: 'var(--accent-blue)' }} />
            <h3 style={{ fontSize: '1.05rem', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
              Operating Delivery Branch
            </h3>
          </div>

          {employee.branch_name ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
              <div>
                Hub Name: <strong style={{ color: 'var(--text-primary)' }}>{employee.branch_name}</strong>
              </div>
              {employee.branch_code && (
                <div>
                  Hub Code: <strong style={{ color: 'var(--accent-cyan)' }}>{employee.branch_code}</strong>
                </div>
              )}
            </div>
          ) : (
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', margin: 0 }}>
              No specific branch assigned. Operating under general tenant jurisdiction.
            </p>
          )}
        </div>

        {/* Assigned Vehicle Card */}
        <div style={{
          background: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px solid var(--border-subtle)',
          padding: '1.5rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <Truck size={22} style={{ color: 'var(--accent-emerald)' }} />
            <h3 style={{ fontSize: '1.05rem', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
              Active Assigned Vehicle
            </h3>
          </div>

          {assigned_vehicle ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
              <div>
                Registration: <strong style={{ color: 'var(--accent-cyan)' }}>{assigned_vehicle.registration_number}</strong>
              </div>
              <div>
                Type: <strong style={{ color: 'var(--text-primary)' }}>{assigned_vehicle.vehicle_type}</strong>
              </div>
              {assigned_vehicle.make_model && (
                <div>
                  Make / Model: <strong style={{ color: 'var(--text-primary)' }}>{assigned_vehicle.make_model}</strong>
                </div>
              )}
              <div style={{ marginTop: '0.25rem' }}>
                <span style={{
                  fontSize: '0.75rem',
                  padding: '0.2rem 0.5rem',
                  borderRadius: '6px',
                  background: 'rgba(16, 185, 129, 0.15)',
                  color: '#34d399',
                  border: '1px solid rgba(16, 185, 129, 0.3)',
                }}>
                  Status: {assigned_vehicle.status}
                </span>
              </div>
            </div>
          ) : (
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', margin: 0 }}>
              No active vehicle assigned currently. Check in with the fleet dispatcher.
            </p>
          )}
        </div>
      </div>

      {/* Profile Details Card */}
      <div style={{
        background: 'var(--bg-card)',
        borderRadius: '12px',
        border: '1px solid var(--border-subtle)',
        padding: '1.5rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '1rem',
      }}>
        <h3 style={{ fontSize: '1.05rem', fontWeight: 600, color: 'var(--text-primary)', margin: 0 }}>
          Employment & Verification Records
        </h3>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1rem', fontSize: '0.85rem' }}>
          {employee.email && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
              <Mail size={16} style={{ color: 'var(--accent-cyan)' }} />
              <span>{employee.email}</span>
            </div>
          )}

          {employee.phone && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
              <Phone size={16} style={{ color: 'var(--accent-emerald)' }} />
              <span>{employee.phone}</span>
            </div>
          )}

          {employee.license_number && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
              <FileBadge size={16} style={{ color: 'var(--accent-amber)' }} />
              <span>Driving Lic: <strong>{employee.license_number}</strong></span>
            </div>
          )}

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
            <ShieldCheck size={16} style={{ color: employee.is_active ? 'var(--accent-emerald)' : 'var(--accent-rose)' }} />
            <span>KYC Verification: <strong>{employee.verification_status}</strong></span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
            <Calendar size={16} style={{ color: 'var(--text-muted)' }} />
            <span>Contract: <strong>{employee.employment_type}</strong></span>
          </div>
        </div>
      </div>
    </div>
  );
};

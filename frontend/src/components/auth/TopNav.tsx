import React, { useState } from 'react';
import { 
  Building2, 
  Users, 
  LogOut, 
  ChevronDown, 
  Boxes 
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { TeamModal } from '../tenants/TeamModal';

export const TopNav: React.FC = () => {
  const { user, tenants, activeTenant, setActiveTenant, logout } = useAuth();
  const [showTeamModal, setShowTeamModal] = useState(false);
  const [showTenantDropdown, setShowTenantDropdown] = useState(false);

  if (!user) return null;

  return (
    <>
      <header style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0.875rem 2rem',
        borderBottom: '1px solid var(--border-subtle)',
        backgroundColor: 'rgba(10, 14, 23, 0.75)',
        backdropFilter: 'blur(12px)',
        position: 'sticky',
        top: 0,
        zIndex: 100
      }}>
        {/* Brand */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '36px',
            height: '36px',
            borderRadius: '10px',
            background: 'linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(59, 130, 246, 0.2))',
            border: '1px solid var(--border-glow)'
          }}>
            <Boxes size={20} color="#06b6d4" />
          </div>
          <div>
            <span style={{ fontWeight: 700, fontSize: '1.1rem', letterSpacing: '-0.025em' }}>
              LogiFlows
            </span>
            <span style={{
              marginLeft: '0.5rem',
              fontSize: '0.65rem',
              fontWeight: 700,
              padding: '0.15rem 0.5rem',
              borderRadius: '6px',
              backgroundColor: 'rgba(6, 182, 212, 0.1)',
              color: 'var(--accent-cyan)',
              border: '1px solid rgba(6, 182, 212, 0.3)'
            }}>
              PHASE 1 MULTI-TENANT
            </span>
          </div>
        </div>

        {/* Tenant Selector & Actions */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          {/* Active Tenant Dropdown */}
          {activeTenant && (
            <div style={{ position: 'relative' }}>
              <button
                type="button"
                onClick={() => setShowTenantDropdown(!showTenantDropdown)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                  padding: '0.45rem 0.85rem',
                  backgroundColor: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  borderRadius: '8px',
                  color: 'var(--text-primary)',
                  fontSize: '0.85rem',
                  cursor: 'pointer'
                }}
              >
                <Building2 size={16} color="#06b6d4" />
                <span style={{ fontWeight: 600 }}>{activeTenant.name}</span>
                <span style={{
                  fontSize: '0.7rem',
                  padding: '0.1rem 0.4rem',
                  borderRadius: '4px',
                  backgroundColor: 'rgba(59, 130, 246, 0.15)',
                  color: '#60a5fa',
                  border: '1px solid rgba(59, 130, 246, 0.3)'
                }}>
                  {activeTenant.role}
                </span>
                {tenants.length > 1 && <ChevronDown size={14} color="var(--text-muted)" />}
              </button>

              {showTenantDropdown && tenants.length > 1 && (
                <div style={{
                  position: 'absolute',
                  top: '100%',
                  right: 0,
                  marginTop: '0.5rem',
                  width: '260px',
                  backgroundColor: 'var(--bg-secondary)',
                  border: '1px solid var(--border-subtle)',
                  borderRadius: '10px',
                  boxShadow: '0 10px 25px rgba(0, 0, 0, 0.5)',
                  padding: '0.5rem',
                  zIndex: 200
                }}>
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', padding: '0.25rem 0.5rem', fontWeight: 600 }}>
                    SWITCH TENANT COMPANY
                  </div>
                  {tenants.map((t) => (
                    <button
                      key={t.id}
                      type="button"
                      onClick={() => {
                        setActiveTenant(t);
                        setShowTenantDropdown(false);
                      }}
                      style={{
                        width: '100%',
                        textAlign: 'left',
                        padding: '0.5rem',
                        backgroundColor: t.id === activeTenant.id ? 'rgba(6, 182, 212, 0.1)' : 'transparent',
                        border: 'none',
                        borderRadius: '6px',
                        color: t.id === activeTenant.id ? 'var(--accent-cyan)' : 'var(--text-primary)',
                        fontSize: '0.85rem',
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between'
                      }}
                    >
                      <span style={{ fontWeight: 500 }}>{t.name}</span>
                      <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>{t.role}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Manage Team Members Button */}
          {activeTenant && (
            <button
              type="button"
              onClick={() => setShowTeamModal(true)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.35rem',
                padding: '0.45rem 0.75rem',
                backgroundColor: 'rgba(15, 23, 42, 0.6)',
                border: '1px solid var(--border-subtle)',
                borderRadius: '8px',
                color: 'var(--text-secondary)',
                fontSize: '0.85rem',
                cursor: 'pointer'
              }}
            >
              <Users size={16} />
              <span>Team</span>
            </button>
          )}

          {/* Current User Badge */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.6rem',
            paddingLeft: '0.5rem',
            borderLeft: '1px solid var(--border-subtle)'
          }}>
            <div style={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              backgroundColor: 'rgba(59, 130, 246, 0.2)',
              border: '1px solid rgba(59, 130, 246, 0.4)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#93c5fd',
              fontSize: '0.8rem',
              fontWeight: 700
            }}>
              {user.full_name ? user.full_name.charAt(0).toUpperCase() : 'U'}
            </div>
            <div>
              <div style={{ fontSize: '0.85rem', fontWeight: 600 }}>{user.full_name}</div>
              <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>{user.email}</div>
            </div>
          </div>

          {/* Logout Button */}
          <button
            type="button"
            onClick={logout}
            title="Sign Out"
            style={{
              background: 'none',
              border: '1px solid var(--border-subtle)',
              borderRadius: '8px',
              color: 'var(--text-muted)',
              cursor: 'pointer',
              padding: '0.45rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
          >
            <LogOut size={16} />
          </button>
        </div>
      </header>

      {/* Team Modal */}
      {showTeamModal && activeTenant && (
        <TeamModal
          tenant={activeTenant}
          onClose={() => setShowTeamModal(false)}
        />
      )}
    </>
  );
};

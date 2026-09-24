import React, { useState, useEffect, useCallback } from 'react';
import { 
  Users, 
  Plus, 
  Search, 
  Phone, 
  Mail, 
  Building, 
  MapPin, 
  Edit3, 
  Trash2, 
  Package, 
  CheckCircle2, 
  AlertCircle, 
  Loader2, 
  FileText,
  X
} from 'lucide-react';
import type { Customer, CreateCustomerPayload, UpdateCustomerPayload, CustomerType } from '../../types/customers';
import type { Parcel } from '../../types/parcels';
import { CustomerModal } from './CustomerModal';
import { api } from '../../services/api';

interface CustomerListProps {
  tenantId: string;
  userRole: string;
  onBookParcelForCustomer?: (customer: Customer) => void;
}

export const CustomerList: React.FC<CustomerListProps> = ({ 
  tenantId, 
  userRole,
  onBookParcelForCustomer,
}) => {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Filters
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  // Modal State
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [modalLoading, setModalLoading] = useState(false);

  // Customer Parcels Viewer
  const [parcelsModalCustomer, setParcelsModalCustomer] = useState<Customer | null>(null);
  const [customerParcels, setCustomerParcels] = useState<Parcel[]>([]);
  const [parcelsLoading, setParcelsLoading] = useState(false);

  const canMutate = ['OWNER', 'ADMIN', 'MANAGER', 'OPERATOR'].includes(userRole);

  const fetchCustomers = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.listCustomers(tenantId, {
        search: searchTerm || undefined,
        customer_type: typeFilter || undefined,
        status: statusFilter || undefined,
        page,
        limit,
      });
      setCustomers(res.customers || []);
      setTotal(res.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch customer directory.');
    } finally {
      setLoading(false);
    }
  }, [tenantId, searchTerm, typeFilter, statusFilter, page, limit]);

  useEffect(() => {
    fetchCustomers();
  }, [fetchCustomers]);

  const handleOpenCreate = () => {
    setSelectedCustomer(null);
    setModalOpen(true);
  };

  const handleOpenEdit = (c: Customer) => {
    setSelectedCustomer(c);
    setModalOpen(true);
  };

  const handleModalSubmit = async (payload: CreateCustomerPayload | UpdateCustomerPayload) => {
    try {
      setModalLoading(true);
      if (selectedCustomer) {
        await api.updateCustomer(tenantId, selectedCustomer.id, payload as UpdateCustomerPayload);
        setSuccess(`Customer profile "${selectedCustomer.customer_code}" updated successfully.`);
      } else {
        const created = await api.createCustomer(tenantId, payload as CreateCustomerPayload);
        setSuccess(`Customer "${created.name}" registered with ID ${created.customer_code}.`);
      }
      setModalOpen(false);
      fetchCustomers();
      setTimeout(() => setSuccess(null), 4000);
    } finally {
      setModalLoading(false);
    }
  };

  const handleDelete = async (customer: Customer) => {
    if (!window.confirm(`Are you sure you want to deactivate customer ${customer.name} (${customer.customer_code})?`)) {
      return;
    }
    try {
      await api.deleteCustomer(tenantId, customer.id);
      setSuccess(`Customer ${customer.customer_code} deactivated.`);
      fetchCustomers();
      setTimeout(() => setSuccess(null), 4000);
    } catch (err: any) {
      setError(err.message || 'Failed to delete customer.');
    }
  };

  const handleViewParcels = async (c: Customer) => {
    setParcelsModalCustomer(c);
    setParcelsLoading(true);
    try {
      const res = await api.getCustomerParcels(tenantId, c.id);
      setCustomerParcels(res.parcels || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load customer parcels.');
    } finally {
      setParcelsLoading(false);
    }
  };

  const getTypeBadgeStyle = (type: CustomerType) => {
    switch (type) {
      case 'ENTERPRISE':
        return { bg: 'rgba(168, 85, 247, 0.15)', text: '#c084fc', border: 'rgba(168, 85, 247, 0.3)' };
      case 'BUSINESS':
        return { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)' };
      case 'MERCHANT':
        return { bg: 'rgba(234, 179, 8, 0.15)', text: '#fde047', border: 'rgba(234, 179, 8, 0.3)' };
      default:
        return { bg: 'rgba(6, 182, 212, 0.15)', text: '#22d3ee', border: 'rgba(6, 182, 212, 0.3)' };
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Header bar */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexWrap: 'wrap',
        gap: '1rem',
      }}>
        <div>
          <h1 style={{
            fontSize: '1.75rem',
            fontWeight: 700,
            margin: 0,
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            color: '#f8fafc',
          }}>
            <Users size={28} color="var(--accent-cyan, #06b6d4)" />
            Customer Management & CRM
          </h1>
          <p style={{ margin: '0.25rem 0 0 0', color: '#94a3b8', fontSize: '0.9rem' }}>
            Consignor profiles, enterprise address book, and customer-linked parcel history
          </p>
        </div>

        {canMutate && (
          <button
            onClick={handleOpenCreate}
            style={{
              padding: '0.65rem 1.25rem',
              borderRadius: '8px',
              border: 'none',
              background: 'linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%)',
              color: '#ffffff',
              fontWeight: 600,
              fontSize: '0.9rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              boxShadow: '0 4px 12px rgba(6, 182, 212, 0.25)',
              transition: 'transform 0.15s ease',
            }}
          >
            <Plus size={18} />
            Register Customer
          </button>
        )}
      </div>

      {/* Notifications */}
      {success && (
        <div style={{
          padding: '0.85rem 1.25rem',
          borderRadius: '8px',
          background: 'rgba(16, 185, 129, 0.15)',
          border: '1px solid rgba(16, 185, 129, 0.3)',
          color: '#34d399',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          fontSize: '0.9rem',
        }}>
          <CheckCircle2 size={18} />
          <span>{success}</span>
        </div>
      )}

      {error && (
        <div style={{
          padding: '0.85rem 1.25rem',
          borderRadius: '8px',
          background: 'rgba(239, 68, 68, 0.15)',
          border: '1px solid rgba(239, 68, 68, 0.3)',
          color: '#f87171',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          fontSize: '0.9rem',
        }}>
          <AlertCircle size={18} />
          <span>{error}</span>
        </div>
      )}

      {/* Filter and Search Bar */}
      <div style={{
        background: 'var(--bg-card, #1e293b)',
        border: '1px solid var(--border-subtle, #334155)',
        borderRadius: '12px',
        padding: '1rem 1.25rem',
        display: 'flex',
        flexWrap: 'wrap',
        alignItems: 'center',
        gap: '1rem',
      }}>
        <div style={{ position: 'relative', flex: '1 1 240px' }}>
          <Search size={16} style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }} />
          <input
            type="text"
            placeholder="Search by name, code, phone, company, or tax ID..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setPage(1);
            }}
            style={{
              width: '100%',
              padding: '0.55rem 0.85rem 0.55rem 2.4rem',
              background: '#0f172a',
              border: '1px solid #334155',
              borderRadius: '8px',
              color: '#f8fafc',
              fontSize: '0.875rem',
              outline: 'none',
              boxSizing: 'border-box',
            }}
          />
        </div>

        <select
          value={typeFilter}
          onChange={(e) => {
            setTypeFilter(e.target.value);
            setPage(1);
          }}
          style={{
            padding: '0.55rem 0.85rem',
            background: '#0f172a',
            border: '1px solid #334155',
            borderRadius: '8px',
            color: '#cbd5e1',
            fontSize: '0.875rem',
            outline: 'none',
          }}
        >
          <option value="">All Classifications</option>
          <option value="INDIVIDUAL">Individual</option>
          <option value="BUSINESS">Business</option>
          <option value="ENTERPRISE">Enterprise</option>
          <option value="MERCHANT">Merchant / E-Com</option>
        </select>

        <select
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value);
            setPage(1);
          }}
          style={{
            padding: '0.55rem 0.85rem',
            background: '#0f172a',
            border: '1px solid #334155',
            borderRadius: '8px',
            color: '#cbd5e1',
            fontSize: '0.875rem',
            outline: 'none',
          }}
        >
          <option value="">All Statuses</option>
          <option value="ACTIVE">Active</option>
          <option value="INACTIVE">Inactive</option>
          <option value="SUSPENDED">Suspended</option>
        </select>

        <span style={{ fontSize: '0.85rem', color: '#94a3b8', marginLeft: 'auto' }}>
          {total} customer{total === 1 ? '' : 's'} registered
        </span>
      </div>

      {/* Directory Table / Cards */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem 0', color: '#94a3b8' }}>
          <Loader2 size={32} className="spin-slow" style={{ margin: '0 auto 1rem', color: 'var(--accent-cyan, #06b6d4)' }} />
          <p>Loading customer profiles...</p>
        </div>
      ) : customers.length === 0 ? (
        <div style={{
          background: 'var(--bg-card, #1e293b)',
          border: '1px dashed var(--border-subtle, #334155)',
          borderRadius: '16px',
          padding: '4rem 2rem',
          textAlign: 'center',
          color: '#94a3b8',
        }}>
          <Users size={48} style={{ margin: '0 auto 1rem', color: '#475569' }} />
          <h3 style={{ margin: 0, fontSize: '1.1rem', color: '#f8fafc' }}>No customers found</h3>
          <p style={{ margin: '0.5rem 0 1.5rem', fontSize: '0.9rem' }}>
            {searchTerm || typeFilter || statusFilter 
              ? 'Try adjusting your search criteria or clearing filters.' 
              : 'Start by registering your first customer profile for easy parcel booking.'}
          </p>
          {canMutate && (
            <button
              onClick={handleOpenCreate}
              style={{
                padding: '0.65rem 1.25rem',
                borderRadius: '8px',
                border: 'none',
                background: 'var(--accent-cyan, #06b6d4)',
                color: '#0f172a',
                fontWeight: 600,
                fontSize: '0.875rem',
                cursor: 'pointer',
              }}
            >
              + Register Customer
            </button>
          )}
        </div>
      ) : (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(340px, 1fr))',
          gap: '1.25rem',
        }}>
          {customers.map((c) => {
            const badge = getTypeBadgeStyle(c.customer_type);
            return (
              <div
                key={c.id}
                style={{
                  background: 'var(--bg-card, #1e293b)',
                  border: '1px solid var(--border-subtle, #334155)',
                  borderRadius: '14px',
                  padding: '1.25rem',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.85rem',
                  boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.2)',
                  transition: 'border-color 0.2s ease, transform 0.2s ease',
                }}
              >
                {/* Card Top: Code & Badges */}
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <span style={{
                    fontFamily: 'monospace',
                    fontSize: '0.8rem',
                    fontWeight: 700,
                    color: 'var(--accent-cyan, #06b6d4)',
                    background: 'rgba(6, 182, 212, 0.1)',
                    padding: '0.2rem 0.5rem',
                    borderRadius: '6px',
                  }}>
                    {c.customer_code}
                  </span>
                  <div style={{ display: 'flex', gap: '0.4rem', alignItems: 'center' }}>
                    <span style={{
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      padding: '0.2rem 0.5rem',
                      borderRadius: '6px',
                      background: badge.bg,
                      color: badge.text,
                      border: `1px solid ${badge.border}`,
                    }}>
                      {c.customer_type}
                    </span>
                    <span style={{
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      padding: '0.2rem 0.5rem',
                      borderRadius: '6px',
                      background: c.status === 'ACTIVE' ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                      color: c.status === 'ACTIVE' ? '#34d399' : '#f87171',
                    }}>
                      {c.status}
                    </span>
                  </div>
                </div>

                {/* Name & Company */}
                <div>
                  <h3 style={{ margin: 0, fontSize: '1.1rem', fontWeight: 600, color: '#f8fafc' }}>
                    {c.name}
                  </h3>
                  {c.company_name && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem', color: '#94a3b8', fontSize: '0.85rem', marginTop: '0.25rem' }}>
                      <Building size={14} />
                      <span>{c.company_name}</span>
                    </div>
                  )}
                </div>

                {/* Contact info */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', fontSize: '0.85rem', color: '#cbd5e1' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <Phone size={14} color="#94a3b8" />
                    <span>{c.phone}</span>
                  </div>
                  {c.email && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <Mail size={14} color="#94a3b8" />
                      <span style={{ textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }}>{c.email}</span>
                    </div>
                  )}
                  {c.tax_id && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#94a3b8', fontSize: '0.8rem' }}>
                      <FileText size={14} />
                      <span>GST/Tax: {c.tax_id}</span>
                    </div>
                  )}
                  <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.5rem', marginTop: '0.2rem' }}>
                    <MapPin size={14} color="#94a3b8" style={{ marginTop: '0.2rem', flexShrink: 0 }} />
                    <span style={{ fontSize: '0.8rem', color: '#94a3b8', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                      {c.billing_address}
                    </span>
                  </div>
                </div>

                {/* Card Actions */}
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  paddingTop: '0.75rem',
                  borderTop: '1px solid #334155',
                  marginTop: 'auto',
                }}>
                  <button
                    onClick={() => handleViewParcels(c)}
                    style={{
                      background: 'none',
                      border: 'none',
                      color: 'var(--accent-cyan, #06b6d4)',
                      fontSize: '0.825rem',
                      fontWeight: 600,
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '0.35rem',
                      padding: 0,
                    }}
                  >
                    <Package size={14} />
                    View Bookings
                  </button>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    {onBookParcelForCustomer && (
                      <button
                        onClick={() => onBookParcelForCustomer(c)}
                        title="Book parcel with this customer"
                        style={{
                          background: 'rgba(6, 182, 212, 0.1)',
                          border: '1px solid rgba(6, 182, 212, 0.3)',
                          borderRadius: '6px',
                          color: 'var(--accent-cyan, #06b6d4)',
                          fontSize: '0.8rem',
                          fontWeight: 500,
                          cursor: 'pointer',
                          padding: '0.3rem 0.6rem',
                        }}
                      >
                        + Book Parcel
                      </button>
                    )}

                    {canMutate && (
                      <>
                        <button
                          onClick={() => handleOpenEdit(c)}
                          title="Edit profile"
                          style={{
                            background: 'none',
                            border: 'none',
                            color: '#94a3b8',
                            cursor: 'pointer',
                            padding: '0.35rem',
                            borderRadius: '6px',
                          }}
                        >
                          <Edit3 size={15} />
                        </button>
                        <button
                          onClick={() => handleDelete(c)}
                          title="Deactivate customer"
                          style={{
                            background: 'none',
                            border: 'none',
                            color: '#ef4444',
                            cursor: 'pointer',
                            padding: '0.35rem',
                            borderRadius: '6px',
                          }}
                        >
                          <Trash2 size={15} />
                        </button>
                      </>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Pagination */}
      {total > limit && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: '0.5rem', marginTop: '1rem' }}>
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            style={{
              padding: '0.5rem 1rem',
              borderRadius: '6px',
              border: '1px solid #334155',
              background: '#0f172a',
              color: page === 1 ? '#64748b' : '#cbd5e1',
              cursor: page === 1 ? 'not-allowed' : 'pointer',
            }}
          >
            Previous
          </button>
          <span style={{ display: 'flex', alignItems: 'center', padding: '0 0.75rem', fontSize: '0.875rem', color: '#94a3b8' }}>
            Page {page} of {Math.ceil(total / limit)}
          </span>
          <button
            onClick={() => setPage((p) => p + 1)}
            disabled={page >= Math.ceil(total / limit)}
            style={{
              padding: '0.5rem 1rem',
              borderRadius: '6px',
              border: '1px solid #334155',
              background: '#0f172a',
              color: page >= Math.ceil(total / limit) ? '#64748b' : '#cbd5e1',
              cursor: page >= Math.ceil(total / limit) ? 'not-allowed' : 'pointer',
            }}
          >
            Next
          </button>
        </div>
      )}

      {/* Register/Edit Customer Modal */}
      <CustomerModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleModalSubmit}
        customer={selectedCustomer}
        loading={modalLoading}
      />

      {/* Customer Linked Parcels Drawer / Modal */}
      {parcelsModalCustomer && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(10, 15, 30, 0.75)',
          backdropFilter: 'blur(6px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000,
          padding: '1.5rem',
        }}>
          <div style={{
            background: 'var(--bg-card, #1e293b)',
            border: '1px solid var(--border-subtle, #334155)',
            borderRadius: '16px',
            width: '100%',
            maxWidth: '780px',
            maxHeight: '85vh',
            display: 'flex',
            flexDirection: 'column',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
            color: '#f8fafc',
          }}>
            {/* Header */}
            <div style={{
              padding: '1.25rem 1.5rem',
              borderBottom: '1px solid #334155',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}>
              <div>
                <h2 style={{ fontSize: '1.15rem', fontWeight: 600, margin: 0, color: 'var(--accent-cyan, #06b6d4)' }}>
                  Parcel Bookings: {parcelsModalCustomer.name}
                </h2>
                <span style={{ fontSize: '0.8rem', color: '#94a3b8', fontFamily: 'monospace' }}>
                  {parcelsModalCustomer.customer_code} • {parcelsModalCustomer.phone}
                </span>
              </div>
              <button
                onClick={() => setParcelsModalCustomer(null)}
                style={{
                  background: 'none',
                  border: 'none',
                  color: '#94a3b8',
                  cursor: 'pointer',
                  padding: '0.5rem',
                }}
              >
                <X size={20} />
              </button>
            </div>

            {/* List */}
            <div style={{ padding: '1.5rem', overflowY: 'auto', flex: 1 }}>
              {parcelsLoading ? (
                <div style={{ textAlign: 'center', padding: '2rem 0', color: '#94a3b8' }}>
                  <Loader2 size={24} className="spin-slow" style={{ margin: '0 auto 0.5rem', color: 'var(--accent-cyan, #06b6d4)' }} />
                  <p>Fetching booking records...</p>
                </div>
              ) : customerParcels.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '3rem 1rem', color: '#94a3b8' }}>
                  <Package size={36} style={{ margin: '0 auto 0.5rem', color: '#475569' }} />
                  <p>No parcels linked to this customer account yet.</p>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  {customerParcels.map((p) => (
                    <div
                      key={p.id}
                      style={{
                        padding: '1rem',
                        background: '#0f172a',
                        border: '1px solid #334155',
                        borderRadius: '10px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        gap: '1rem',
                      }}
                    >
                      <div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                          <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#f8fafc', fontSize: '0.9rem' }}>
                            {p.tracking_number}
                          </span>
                          <span style={{
                            fontSize: '0.75rem',
                            padding: '0.15rem 0.45rem',
                            borderRadius: '4px',
                            background: 'rgba(6, 182, 212, 0.15)',
                            color: 'var(--accent-cyan, #06b6d4)',
                            fontWeight: 600,
                          }}>
                            {p.status}
                          </span>
                          <span style={{ fontSize: '0.75rem', color: '#94a3b8' }}>
                            {p.service_type}
                          </span>
                        </div>
                        <div style={{ fontSize: '0.8rem', color: '#94a3b8', marginTop: '0.25rem' }}>
                          {p.sender_name} ➔ {p.receiver_name} • {p.weight_kg} kg
                          {p.price !== undefined && p.price !== null && (
                            <span style={{ marginLeft: '0.5rem', color: '#34d399', fontWeight: 600 }}>
                              ₹{p.price.toFixed(2)}
                            </span>
                          )}
                        </div>
                      </div>

                      <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
                        {new Date(p.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

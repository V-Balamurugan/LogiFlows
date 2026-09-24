import React, { useState } from 'react';
import { X, User, Building, Phone, Mail, MapPin, FileText, AlertCircle, Loader2 } from 'lucide-react';
import type { Customer, CreateCustomerPayload, UpdateCustomerPayload, CustomerType } from '../../types/customers';

interface CustomerModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (payload: CreateCustomerPayload | UpdateCustomerPayload) => Promise<void>;
  customer?: Customer | null;
  loading: boolean;
}

export const CustomerModal: React.FC<CustomerModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  customer,
  loading,
}) => {
  const [customerType, setCustomerType] = useState<CustomerType>(customer?.customer_type || 'INDIVIDUAL');
  const [name, setName] = useState(customer?.name || '');
  const [companyName, setCompanyName] = useState(customer?.company_name || '');
  const [phone, setPhone] = useState(customer?.phone || '');
  const [email, setEmail] = useState(customer?.email || '');
  const [taxId, setTaxId] = useState(customer?.tax_id || '');
  const [billingAddress, setBillingAddress] = useState(customer?.billing_address || '');
  const [shippingAddress, setShippingAddress] = useState(customer?.shipping_address || '');
  const [notes, setNotes] = useState(customer?.notes || '');
  const [status, setStatus] = useState<Customer['status']>(customer?.status || 'ACTIVE');
  const [validationError, setValidationError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    if (!name.trim()) {
      setValidationError('Customer name is required.');
      return;
    }
    if (!phone.trim() || phone.trim().length < 7) {
      setValidationError('A valid contact phone number (min 7 digits) is required.');
      return;
    }
    if (!billingAddress.trim()) {
      setValidationError('Billing address is required for invoices and consignments.');
      return;
    }

    try {
      if (customer) {
        await onSubmit({
          customer_type: customerType,
          name: name.trim(),
          company_name: companyName.trim() || undefined,
          phone: phone.trim(),
          email: email.trim() || undefined,
          tax_id: taxId.trim() || undefined,
          billing_address: billingAddress.trim(),
          shipping_address: shippingAddress.trim() || undefined,
          status,
          notes: notes.trim() || undefined,
        });
      } else {
        await onSubmit({
          customer_type: customerType,
          name: name.trim(),
          company_name: companyName.trim() || undefined,
          phone: phone.trim(),
          email: email.trim() || undefined,
          tax_id: taxId.trim() || undefined,
          billing_address: billingAddress.trim(),
          shipping_address: shippingAddress.trim() || undefined,
          notes: notes.trim() || undefined,
        });
      }
    } catch (err: any) {
      setValidationError(err.message || 'Failed to save customer profile.');
    }
  };

  return (
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
        maxWidth: '680px',
        maxHeight: '90vh',
        overflowY: 'auto',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
        color: '#f8fafc',
      }}>
        {/* Header */}
        <div style={{
          padding: '1.5rem 2rem',
          borderBottom: '1px solid var(--border-subtle, #334155)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}>
          <div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600, margin: 0, color: 'var(--accent-cyan, #06b6d4)' }}>
              {customer ? 'Update Customer Profile' : 'Register New Customer'}
            </h2>
            <p style={{ margin: '0.25rem 0 0 0', fontSize: '0.85rem', color: '#94a3b8' }}>
              {customer ? `Editing profile for ${customer.customer_code}` : 'Onboard individual or business consignor/consignee for parcels'}
            </p>
          </div>
          <button
            onClick={onClose}
            disabled={loading}
            style={{
              background: 'none',
              border: 'none',
              color: '#94a3b8',
              cursor: 'pointer',
              padding: '0.5rem',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} style={{ padding: '2rem' }}>
          {validationError && (
            <div style={{
              background: 'rgba(239, 68, 68, 0.15)',
              border: '1px solid rgba(239, 68, 68, 0.3)',
              borderRadius: '8px',
              padding: '0.75rem 1rem',
              marginBottom: '1.5rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              color: '#f87171',
              fontSize: '0.875rem',
            }}>
              <AlertCircle size={18} />
              <span>{validationError}</span>
            </div>
          )}

          {/* Customer Type Selector */}
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.5rem' }}>
              Customer Classification *
            </label>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '0.5rem' }}>
              {(['INDIVIDUAL', 'BUSINESS', 'ENTERPRISE', 'MERCHANT'] as CustomerType[]).map((type) => (
                <button
                  type="button"
                  key={type}
                  onClick={() => setCustomerType(type)}
                  style={{
                    padding: '0.6rem 0.5rem',
                    borderRadius: '8px',
                    border: customerType === type ? '2px solid var(--accent-cyan, #06b6d4)' : '1px solid #334155',
                    background: customerType === type ? 'rgba(6, 182, 212, 0.15)' : 'rgba(15, 23, 42, 0.4)',
                    color: customerType === type ? 'var(--accent-cyan, #06b6d4)' : '#cbd5e1',
                    fontSize: '0.8rem',
                    fontWeight: customerType === type ? 600 : 500,
                    cursor: 'pointer',
                    transition: 'all 0.15s ease',
                  }}
                >
                  {type}
                </button>
              ))}
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
            {/* Name */}
            <div>
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
                Full Name / Contact Person *
              </label>
              <div style={{ position: 'relative' }}>
                <User size={16} style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }} />
                <input
                  type="text"
                  required
                  placeholder="e.g. Priya Sharma"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                />
              </div>
            </div>

            {/* Company Name */}
            <div>
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
                Company / Organization {customerType !== 'INDIVIDUAL' && '*'}
              </label>
              <div style={{ position: 'relative' }}>
                <Building size={16} style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }} />
                <input
                  type="text"
                  placeholder={customerType === 'INDIVIDUAL' ? 'Optional' : 'e.g. Tata Retail Pvt Ltd'}
                  value={companyName}
                  onChange={(e) => setCompanyName(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                />
              </div>
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
            {/* Phone */}
            <div>
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
                Primary Phone *
              </label>
              <div style={{ position: 'relative' }}>
                <Phone size={16} style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }} />
                <input
                  type="tel"
                  required
                  placeholder="+91 98765 43210"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                />
              </div>
            </div>

            {/* Email */}
            <div>
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
                Email Address
              </label>
              <div style={{ position: 'relative' }}>
                <Mail size={16} style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }} />
                <input
                  type="email"
                  placeholder="contact@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                />
              </div>
            </div>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
            {/* Tax ID */}
            <div>
              <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
                Tax / GSTIN Number
              </label>
              <div style={{ position: 'relative' }}>
                <FileText size={16} style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)', color: '#64748b' }} />
                <input
                  type="text"
                  placeholder="e.g. 33AAAAA0000A1Z5"
                  value={taxId}
                  onChange={(e) => setTaxId(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                />
              </div>
            </div>

            {/* Status (if editing) */}
            {customer && (
              <div>
                <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
                  Account Status
                </label>
                <select
                  value={status}
                  onChange={(e) => setStatus(e.target.value as Customer['status'])}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                >
                  <option value="ACTIVE">ACTIVE</option>
                  <option value="INACTIVE">INACTIVE</option>
                  <option value="SUSPENDED">SUSPENDED</option>
                </select>
              </div>
            )}
          </div>

          {/* Billing Address */}
          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
              Billing / Registered Address *
            </label>
            <div style={{ position: 'relative' }}>
              <MapPin size={16} style={{ position: 'absolute', left: '0.85rem', top: '0.85rem', color: '#64748b' }} />
              <textarea
                required
                rows={2}
                placeholder="Street address, City, State, PIN Code"
                value={billingAddress}
                onChange={(e) => setBillingAddress(e.target.value)}
                style={{
                  width: '100%',
                  padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                  background: '#0f172a',
                  border: '1px solid #334155',
                  borderRadius: '8px',
                  color: '#f8fafc',
                  fontSize: '0.9rem',
                  outline: 'none',
                  boxSizing: 'border-box',
                  resize: 'vertical',
                }}
              />
            </div>
          </div>

          {/* Shipping Address */}
          <div style={{ marginBottom: '1rem' }}>
            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
              Default Delivery / Shipping Address (Optional)
            </label>
            <div style={{ position: 'relative' }}>
              <MapPin size={16} style={{ position: 'absolute', left: '0.85rem', top: '0.85rem', color: '#64748b' }} />
              <textarea
                rows={2}
                placeholder="Warehouse or secondary delivery destination (if different from billing)"
                value={shippingAddress}
                onChange={(e) => setShippingAddress(e.target.value)}
                style={{
                  width: '100%',
                  padding: '0.65rem 0.85rem 0.65rem 2.4rem',
                  background: '#0f172a',
                  border: '1px solid #334155',
                  borderRadius: '8px',
                  color: '#f8fafc',
                  fontSize: '0.9rem',
                  outline: 'none',
                  boxSizing: 'border-box',
                  resize: 'vertical',
                }}
              />
            </div>
          </div>

          {/* Internal Notes */}
          <div style={{ marginBottom: '1.5rem' }}>
            <label style={{ display: 'block', fontSize: '0.85rem', fontWeight: 500, color: '#94a3b8', marginBottom: '0.4rem' }}>
              Logistics Notes & Special Instructions
            </label>
            <input
              type="text"
              placeholder="e.g. VIP Enterprise client; priority pickup before 10 AM"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              style={{
                width: '100%',
                padding: '0.65rem 0.85rem',
                background: '#0f172a',
                border: '1px solid #334155',
                borderRadius: '8px',
                color: '#f8fafc',
                fontSize: '0.9rem',
                outline: 'none',
                boxSizing: 'border-box',
              }}
            />
          </div>

          {/* Actions */}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              style={{
                padding: '0.65rem 1.25rem',
                borderRadius: '8px',
                border: '1px solid #334155',
                background: 'transparent',
                color: '#94a3b8',
                fontWeight: 500,
                fontSize: '0.9rem',
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              style={{
                padding: '0.65rem 1.5rem',
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
              }}
            >
              {loading && <Loader2 size={16} className="spin-slow" />}
              {customer ? 'Save Changes' : 'Complete Registration'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

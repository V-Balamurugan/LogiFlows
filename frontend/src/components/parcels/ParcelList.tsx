import React, { useState, useEffect } from 'react';
import { 
  Package, 
  Search, 
  Filter, 
  Plus, 
  QrCode, 
  ArrowRight, 
  Clock, 
  AlertCircle, 
  CheckCircle2, 
  RefreshCw, 
  X,
  MapPin,
  Scan,
  Printer,
  Building2,
  Loader2
} from 'lucide-react';
import { api } from '../../services/api';
import { useAuth } from '../../context/AuthContext';
import type { 
  Parcel, 
  ParcelStatus, 
  ServiceType, 
  CreateParcelPayload, 
  ParcelStatusHistory 
} from '../../types/parcels';
import type { Branch } from '../../types/resources';
import type { Customer } from '../../types/customers';
import { ParcelLabelModal } from './ParcelLabelModal';

interface ParcelListProps {
  tenantId: string;
  userRole?: string;
}

const VALID_NEXT_TRANSITIONS: Record<ParcelStatus, ParcelStatus[]> = {
  CREATED: ['BOOKED', 'RECEIVED_AT_ORIGIN_BRANCH', 'CANCELLED'],
  BOOKED: ['READY_FOR_PICKUP', 'RECEIVED_AT_ORIGIN_BRANCH', 'CANCELLED'],
  READY_FOR_PICKUP: ['PICKED_UP', 'CANCELLED'],
  PICKED_UP: ['RECEIVED_AT_ORIGIN_BRANCH'],
  RECEIVED_AT_ORIGIN_BRANCH: ['IN_TRANSIT', 'OUT_FOR_DELIVERY', 'ON_HOLD'],
  IN_TRANSIT: ['RECEIVED_AT_TRANSFER_BRANCH', 'RECEIVED_AT_ORIGIN_BRANCH', 'ON_HOLD'],
  RECEIVED_AT_TRANSFER_BRANCH: ['OUT_FOR_DELIVERY', 'IN_TRANSIT', 'ON_HOLD'],
  OUT_FOR_DELIVERY: ['DELIVERED', 'DELIVERY_ATTEMPTED', 'DELIVERY_FAILED'],
  DELIVERY_ATTEMPTED: ['OUT_FOR_DELIVERY', 'RETURN_INITIATED', 'ON_HOLD'],
  ON_HOLD: ['RECEIVED_AT_ORIGIN_BRANCH', 'RECEIVED_AT_TRANSFER_BRANCH', 'OUT_FOR_DELIVERY', 'RETURN_INITIATED', 'CANCELLED'],
  DELIVERY_FAILED: ['RETURN_INITIATED'],
  RETURN_INITIATED: ['IN_TRANSIT', 'RETURNED'],
  DELIVERED: [],
  RETURNED: [],
  CANCELLED: [],
};

const STATUS_LABELS: Record<ParcelStatus, string> = {
  CREATED: 'Created (Pending Intake)',
  BOOKED: 'Booked',
  READY_FOR_PICKUP: 'Ready for Pickup',
  PICKED_UP: 'Picked Up by Driver',
  RECEIVED_AT_ORIGIN_BRANCH: 'Received at Origin Hub',
  IN_TRANSIT: 'In Transit between Hubs',
  RECEIVED_AT_TRANSFER_BRANCH: 'Received at Transfer/Delivery Hub',
  OUT_FOR_DELIVERY: 'Out for Final Delivery',
  DELIVERY_ATTEMPTED: 'Delivery Attempted',
  DELIVERED: 'Delivered to Recipient',
  DELIVERY_FAILED: 'Delivery Failed',
  RETURN_INITIATED: 'Return Initiated',
  RETURNED: 'Returned to Sender',
  CANCELLED: 'Cancelled',
  ON_HOLD: 'On Hold / Exception',
};

const STATUS_COLORS: Record<ParcelStatus, { bg: string; text: string; border: string }> = {
  CREATED: { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa', border: 'rgba(59, 130, 246, 0.3)' },
  BOOKED: { bg: 'rgba(99, 102, 241, 0.15)', text: '#818cf8', border: 'rgba(99, 102, 241, 0.3)' },
  READY_FOR_PICKUP: { bg: 'rgba(147, 51, 234, 0.15)', text: '#c084fc', border: 'rgba(147, 51, 234, 0.3)' },
  PICKED_UP: { bg: 'rgba(168, 85, 247, 0.15)', text: '#d8b4fe', border: 'rgba(168, 85, 247, 0.3)' },
  RECEIVED_AT_ORIGIN_BRANCH: { bg: 'rgba(14, 165, 233, 0.15)', text: '#38bdf8', border: 'rgba(14, 165, 233, 0.3)' },
  IN_TRANSIT: { bg: 'rgba(139, 92, 246, 0.2)', text: '#a78bfa', border: 'rgba(139, 92, 246, 0.4)' },
  RECEIVED_AT_TRANSFER_BRANCH: { bg: 'rgba(6, 182, 212, 0.15)', text: '#22d3ee', border: 'rgba(6, 182, 212, 0.3)' },
  OUT_FOR_DELIVERY: { bg: 'rgba(245, 158, 11, 0.18)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.4)' },
  DELIVERY_ATTEMPTED: { bg: 'rgba(249, 115, 22, 0.15)', text: '#fb923c', border: 'rgba(249, 115, 22, 0.3)' },
  DELIVERED: { bg: 'rgba(16, 185, 129, 0.18)', text: '#34d399', border: 'rgba(16, 185, 129, 0.4)' },
  DELIVERY_FAILED: { bg: 'rgba(244, 63, 94, 0.15)', text: '#f43f5e', border: 'rgba(244, 63, 94, 0.3)' },
  RETURN_INITIATED: { bg: 'rgba(236, 72, 153, 0.15)', text: '#f472b6', border: 'rgba(236, 72, 153, 0.3)' },
  RETURNED: { bg: 'rgba(156, 163, 175, 0.15)', text: '#9ca3af', border: 'rgba(156, 163, 175, 0.3)' },
  CANCELLED: { bg: 'rgba(239, 68, 68, 0.15)', text: '#ef4444', border: 'rgba(239, 68, 68, 0.3)' },
  ON_HOLD: { bg: 'rgba(234, 179, 8, 0.15)', text: '#facc15', border: 'rgba(234, 179, 8, 0.3)' },
};

const SERVICE_BADGES: Record<ServiceType, { label: string; color: string }> = {
  STANDARD: { label: 'Standard', color: '#94a3b8' },
  EXPRESS: { label: 'Express (48h)', color: '#38bdf8' },
  OVERNIGHT: { label: 'Overnight (24h)', color: '#a855f7' },
  SAME_DAY: { label: 'Same Day Priority', color: '#f59e0b' },
};

export const ParcelList: React.FC<ParcelListProps> = ({ tenantId, userRole }) => {
  const { tenants } = useAuth();
  const [parcels, setParcels] = useState<Parcel[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [customerList, setCustomerList] = useState<Customer[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Multi-Company/Tenant Booking State
  const [bookingTenantId, setBookingTenantId] = useState<string>(tenantId);
  const [bookingBranches, setBookingBranches] = useState<Branch[]>(branches);
  const [bookingCustomers, setBookingCustomers] = useState<Customer[]>(customerList);
  const [loadingTenantData, setLoadingTenantData] = useState<boolean>(false);

  useEffect(() => {
    setBookingTenantId(tenantId);
  }, [tenantId]);

  useEffect(() => {
    setBookingBranches(branches);
  }, [branches]);

  useEffect(() => {
    setBookingCustomers(customerList);
  }, [customerList]);

  // Filters
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [branchFilter, setBranchFilter] = useState<string>('');

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [showScanModal, setShowScanModal] = useState(false);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const [showLabelModal, setShowLabelModal] = useState(false);
  const [labelParcel, setLabelParcel] = useState<Parcel | null>(null);

  // Selected item state
  const [selectedParcel, setSelectedParcel] = useState<Parcel | null>(null);
  const [timeline, setTimeline] = useState<ParcelStatusHistory[]>([]);
  const [timelineLoading, setTimelineLoading] = useState(false);
  const [qrCodePayload, setQrCodePayload] = useState<string>('');

  // Scanner state
  const [scanInput, setScanInput] = useState('');
  const [scanResult, setScanResult] = useState<any | null>(null);
  const [scanLoading, setScanLoading] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);

  // Status update state
  const [targetStatus, setTargetStatus] = useState<ParcelStatus>('RECEIVED_AT_ORIGIN_BRANCH');
  const [statusNotes, setStatusNotes] = useState('');
  const [statusLoading, setStatusLoading] = useState(false);

  // New parcel form state
  const [formData, setFormData] = useState<CreateParcelPayload>({
    sender_customer_id: '',
    receiver_customer_id: '',
    sender_name: '',
    sender_phone: '',
    sender_email: '',
    sender_address: '',
    receiver_name: '',
    receiver_phone: '',
    receiver_email: '',
    receiver_address: '',
    origin_branch_id: '',
    destination_branch_id: '',
    weight_kg: 1.0,
    dimensions_cm: '30x20x15',
    service_type: 'STANDARD',
    declared_value: 0,
    special_instructions: '',
  });
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  const isOperator = userRole === 'TENANT_ADMIN' || userRole === 'TENANT_OPERATOR' || userRole === 'PLATFORM_ADMIN';

  const getEstimatedPrice = (serviceType: ServiceType, weightKg: number, declaredValue: number = 0) => {
    let base = 50;
    if (serviceType === 'EXPRESS') base = 120;
    else if (serviceType === 'OVERNIGHT') base = 200;
    else if (serviceType === 'SAME_DAY') base = 350;

    const weightFee = Math.max(0, weightKg) * 20;
    let insuranceFee = 0;
    if (declaredValue > 1000) {
      insuranceFee = (declaredValue - 1000) * 0.005;
    }
    return {
      base,
      weightFee,
      insuranceFee,
      total: Math.round((base + weightFee + insuranceFee) * 100) / 100,
    };
  };

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [parcelRes, branchRes, customerRes] = await Promise.all([
        api.listParcels(tenantId, {
          search: searchTerm || undefined,
          status: statusFilter || undefined,
          origin_branch_id: branchFilter || undefined,
        }),
        api.listBranches(tenantId),
        api.listCustomers(tenantId, { limit: 100 }).catch(() => ({ customers: [], total: 0, page: 1, limit: 100 })),
      ]);
      setParcels(parcelRes.parcels || []);
      setBranches(branchRes.branches || []);
      setCustomerList(customerRes.customers || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load parcels catalog');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [tenantId, statusFilter, branchFilter]);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    loadData();
  };

  const handleOpenDetails = async (parcel: Parcel) => {
    setSelectedParcel(parcel);
    setShowDetailModal(true);
    setTimelineLoading(true);
    try {
      const [tList, qrRes] = await Promise.all([
        api.getParcelTimeline(tenantId, parcel.id),
        api.getParcelQRCode(tenantId, parcel.id),
      ]);
      setTimeline(tList || []);
      setQrCodePayload(qrRes.qr_payload);
    } catch (err) {
      console.error('Failed to load parcel timeline or QR:', err);
    } finally {
      setTimelineLoading(false);
    }
  };

  const handleOpenStatusModal = (p: Parcel) => {
    setSelectedParcel(p);
    const valid = VALID_NEXT_TRANSITIONS[p.status] || [];
    if (valid.length > 0) {
      setTargetStatus(valid[0]);
    }
    setStatusNotes('');
    setShowStatusModal(true);
  };

  const handleQuickStatusAdvance = async (parcelId: string, nextStatus: ParcelStatus, notes?: string) => {
    try {
      await api.updateParcelStatus(tenantId, parcelId, nextStatus, notes);
      setSuccessMsg(`Status updated to ${STATUS_LABELS[nextStatus] || nextStatus}!`);
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
      if (labelParcel && labelParcel.id === parcelId) {
        setLabelParcel((prev) => prev ? { ...prev, status: nextStatus } : null);
      }
      if (selectedParcel && selectedParcel.id === parcelId) {
        setSelectedParcel((prev) => prev ? { ...prev, status: nextStatus } : null);
      }
    } catch (err: any) {
      alert(`Status advance failed: ${err.message}`);
      throw err;
    }
  };

  const handleBookingTenantChange = async (newTenantId: string) => {
    setBookingTenantId(newTenantId);
    setFormData((prev) => ({
      ...prev,
      origin_branch_id: '',
      destination_branch_id: '',
      sender_customer_id: '',
      receiver_customer_id: '',
    }));
    setLoadingTenantData(true);
    setCreateError(null);
    try {
      const [bRes, cRes] = await Promise.all([
        api.listBranches(newTenantId),
        api.listCustomers(newTenantId, { limit: 100 }).catch(() => ({ customers: [], total: 0, page: 1, limit: 100 })),
      ]);
      setBookingBranches(bRes.branches || []);
      setBookingCustomers(cRes.customers || []);
    } catch (err: any) {
      setCreateError(`Failed to load branches and customers for selected company: ${err.message}`);
    } finally {
      setLoadingTenantData(false);
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.origin_branch_id || !formData.destination_branch_id) {
      setCreateError('Please select both origin and destination hub branches');
      return;
    }
    if (formData.origin_branch_id === formData.destination_branch_id) {
      setCreateError('Origin and destination hub branches must be different');
      return;
    }
    if (formData.weight_kg <= 0) {
      setCreateError('Weight must be greater than 0 kg');
      return;
    }

    try {
      setCreateLoading(true);
      setCreateError(null);
      const created = await api.createParcel(bookingTenantId, formData);
      setShowCreateModal(false);
      setLabelParcel(created);
      setShowLabelModal(true);
      const selectedTenantName = tenants.find((t) => t.id === bookingTenantId)?.name || 'Company';
      setSuccessMsg(`Parcel ${created.tracking_number} registered under ${selectedTenantName}! QR code & shipping label generated.`);
      setTimeout(() => setSuccessMsg(null), 5000);
      if (bookingTenantId === tenantId) {
        loadData();
      }
    } catch (err: any) {
      setCreateError(err.message || 'Failed to create parcel');
    } finally {
      setCreateLoading(false);
    }
  };

  const handleStatusSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedParcel) return;
    try {
      setStatusLoading(true);
      await api.updateParcelStatus(tenantId, selectedParcel.id, targetStatus, statusNotes || undefined);
      setShowStatusModal(false);
      setSuccessMsg(`Status for ${selectedParcel.tracking_number} updated to ${targetStatus}`);
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      alert(`Status update failed: ${err.message}`);
    } finally {
      setStatusLoading(false);
    }
  };

  const handleScanSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!scanInput.trim()) return;
    try {
      setScanLoading(true);
      setScanError(null);
      setScanResult(null);

      // Support either scanning raw QR JSON or typing standard tracking number
      const isJson = scanInput.trim().startsWith('{');
      const payload = isJson 
        ? { qr_payload: scanInput.trim() } 
        : { tracking_number: scanInput.trim().toUpperCase() };

      const res = await api.scanParcel(tenantId, payload);
      setScanResult(res);
      loadData();
    } catch (err: any) {
      setScanError(err.message || 'Scan verification failed. Code may be invalid or belongs to another tenant.');
    } finally {
      setScanLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Header section */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <Package size={26} color="var(--accent-cyan)" />
            Parcels & Cargo Lifecycle
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginTop: '0.2rem' }}>
            Intake registration, immutable custody tracking, automated QR codes, and multi-hub state workflows.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <button
            onClick={() => {
              setScanInput('');
              setScanResult(null);
              setScanError(null);
              setShowScanModal(true);
            }}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.6rem 1rem',
              borderRadius: '8px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(30, 41, 59, 0.7)',
              color: 'var(--text-primary)',
              cursor: 'pointer',
              fontWeight: 500,
            }}
          >
            <Scan size={17} color="var(--accent-purple)" />
            Scan QR / Barcode
          </button>

          {isOperator && (
            <button
              onClick={() => {
                setCreateError(null);
                setShowCreateModal(true);
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                padding: '0.6rem 1.2rem',
                borderRadius: '8px',
                border: 'none',
                background: 'linear-gradient(135deg, var(--accent-cyan) 0%, #2563eb 100%)',
                color: '#fff',
                cursor: 'pointer',
                fontWeight: 600,
                boxShadow: '0 4px 12px rgba(6, 182, 212, 0.25)',
              }}
            >
              <Plus size={18} />
              Book New Parcel
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

      {/* Filter and Search Bar */}
      <div className="glass-panel" style={{ padding: '1rem', display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
        <form onSubmit={handleSearchSubmit} style={{ display: 'flex', flex: 1, minWidth: '240px', gap: '0.5rem' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <Search size={16} style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
            <input
              type="text"
              placeholder="Search by tracking number, sender, or recipient..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              style={{
                width: '100%',
                padding: '0.55rem 0.75rem 0.55rem 2.2rem',
                borderRadius: '6px',
                border: '1px solid var(--border-subtle)',
                background: 'rgba(15, 23, 42, 0.6)',
                color: 'var(--text-primary)',
                fontSize: '0.9rem',
              }}
            />
          </div>
          <button type="submit" style={{ padding: '0.55rem 1rem', borderRadius: '6px', border: 'none', background: 'var(--accent-cyan)', color: '#0f172a', fontWeight: 600, cursor: 'pointer' }}>
            Filter
          </button>
        </form>

        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <Filter size={15} color="var(--text-muted)" />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.8rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Lifecycle States</option>
            {Object.keys(STATUS_COLORS).map((st) => (
              <option key={st} value={st}>{st.replace(/_/g, ' ')}</option>
            ))}
          </select>

          <select
            value={branchFilter}
            onChange={(e) => setBranchFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.8rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Branches</option>
            {branches.map((b) => (
              <option key={b.id} value={b.id}>{b.name} ({b.city})</option>
            ))}
          </select>
        </div>
      </div>

      {/* Main Table View */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
          <RefreshCw className="spin" size={24} style={{ marginBottom: '0.5rem' }} />
          <p>Loading cargo manifests...</p>
        </div>
      ) : error ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--accent-rose)' }}>
          <AlertCircle size={28} style={{ margin: '0 auto 0.5rem auto' }} />
          <p>{error}</p>
        </div>
      ) : parcels.length === 0 ? (
        <div className="glass-panel" style={{ padding: '4rem 2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          <Package size={44} style={{ margin: '0 auto 1rem auto', opacity: 0.4 }} />
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-secondary)' }}>No parcels found</h3>
          <p style={{ fontSize: '0.85rem', marginTop: '0.25rem' }}>
            {searchTerm || statusFilter ? 'Try clearing your active filters.' : 'Click "Book New Parcel" to initiate cargo intake.'}
          </p>
        </div>
      ) : (
        <div className="glass-panel" style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem', textAlign: 'left' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>
                <th style={{ padding: '1rem' }}>Tracking Code</th>
                <th style={{ padding: '1rem' }}>Service & Cargo</th>
                <th style={{ padding: '1rem' }}>Routing (Origin &rarr; Dest)</th>
                <th style={{ padding: '1rem' }}>Parties</th>
                <th style={{ padding: '1rem' }}>Current Status</th>
                <th style={{ padding: '1rem', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {parcels.map((p) => {
                const st = STATUS_COLORS[p.status] || { bg: 'rgba(255,255,255,0.1)', text: '#fff', border: 'transparent' };
                const srv = SERVICE_BADGES[p.service_type] || { label: p.service_type, color: '#aaa' };
                return (
                  <tr key={p.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)', transition: 'background 0.2s' }}>
                    <td style={{ padding: '1rem' }}>
                      <div style={{ fontWeight: 700, fontFamily: 'monospace', color: 'var(--accent-cyan)' }}>
                        {p.tracking_number}
                      </div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                        {new Date(p.created_at).toLocaleDateString()} {new Date(p.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </div>
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <span style={{
                        display: 'inline-block',
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        background: 'rgba(255,255,255,0.06)',
                        color: srv.color,
                        marginBottom: '4px'
                      }}>
                        {srv.label}
                      </span>
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                        {p.weight_kg} kg • {p.dimensions_cm}
                      </div>
                      {p.price !== undefined && p.price !== null && (
                        <div style={{ fontSize: '0.8rem', color: '#34d399', fontWeight: 600, marginTop: '2px' }}>
                          ₹{p.price.toFixed(2)}
                        </div>
                      )}
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
                        <span>{p.origin_branch_name || 'Origin Hub'}</span>
                        <ArrowRight size={13} color="var(--text-muted)" />
                        <span style={{ fontWeight: 600 }}>{p.destination_branch_name || 'Delivery Hub'}</span>
                      </div>
                      {p.current_branch_name && (
                        <div style={{ fontSize: '0.75rem', color: 'var(--accent-purple)', marginTop: '3px', display: 'flex', alignItems: 'center', gap: '3px' }}>
                          <MapPin size={11} /> Current: {p.current_branch_name}
                        </div>
                      )}
                    </td>

                    <td style={{ padding: '1rem', fontSize: '0.85rem' }}>
                      <div><span style={{ color: 'var(--text-muted)' }}>From:</span> {p.sender_name}</div>
                      <div><span style={{ color: 'var(--text-muted)' }}>To:</span> {p.receiver_name}</div>
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
                        {p.status.replace(/_/g, ' ')}
                      </span>
                    </td>

                    <td style={{ padding: '1rem', textAlign: 'right' }}>
                      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.4rem', alignItems: 'center' }}>
                        {/* Direct QR & Label Download Button */}
                        <button
                          onClick={() => {
                            setLabelParcel(p);
                            setShowLabelModal(true);
                          }}
                          title="Generate QR code & print/download shipping label with destination and Parcel ID"
                          style={{
                            padding: '0.45rem 0.75rem',
                            borderRadius: '6px',
                            border: '1px solid rgba(56, 189, 248, 0.35)',
                            background: 'rgba(56, 189, 248, 0.12)',
                            color: 'var(--accent-cyan)',
                            fontSize: '0.8rem',
                            fontWeight: 600,
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '4px'
                          }}
                        >
                          <QrCode size={14} />
                          Label & QR
                        </button>

                        <button
                          onClick={() => handleOpenDetails(p)}
                          style={{
                            padding: '0.45rem 0.75rem',
                            borderRadius: '6px',
                            border: '1px solid var(--border-subtle)',
                            background: 'rgba(30, 41, 59, 0.7)',
                            color: 'var(--text-primary)',
                            fontSize: '0.8rem',
                            cursor: 'pointer',
                          }}
                        >
                          Details
                        </button>

                        {/* Quick 1-Click Lifecycle Advance */}
                        {isOperator && (p.status === 'CREATED' || p.status === 'BOOKED') && (
                          <button
                            onClick={() => handleQuickStatusAdvance(p.id, 'RECEIVED_AT_ORIGIN_BRANCH', 'Hub intake confirmation')}
                            title="Intake package at Origin Hub"
                            style={{
                              padding: '0.45rem 0.75rem',
                              borderRadius: '6px',
                              border: 'none',
                              background: 'rgba(16, 185, 129, 0.2)',
                              color: 'var(--accent-emerald)',
                              fontSize: '0.8rem',
                              fontWeight: 700,
                              cursor: 'pointer',
                              display: 'flex',
                              alignItems: 'center',
                              gap: '4px'
                            }}
                          >
                            <CheckCircle2 size={13} />
                            Intake Hub
                          </button>
                        )}

                        {isOperator && p.status === 'IN_TRANSIT' && (
                          <button
                            onClick={() => handleQuickStatusAdvance(p.id, 'RECEIVED_AT_TRANSFER_BRANCH', 'Arrived at delivery hub')}
                            title="Receive cargo at transfer hub"
                            style={{
                              padding: '0.45rem 0.75rem',
                              borderRadius: '6px',
                              border: 'none',
                              background: 'rgba(6, 182, 212, 0.2)',
                              color: '#22d3ee',
                              fontSize: '0.8rem',
                              fontWeight: 700,
                              cursor: 'pointer',
                            }}
                          >
                            Receive Hub
                          </button>
                        )}

                        {isOperator && p.status !== 'DELIVERED' && p.status !== 'CANCELLED' && p.status !== 'RETURNED' && (
                          <button
                            onClick={() => handleOpenStatusModal(p)}
                            style={{
                              padding: '0.45rem 0.75rem',
                              borderRadius: '6px',
                              border: 'none',
                              background: 'rgba(59, 130, 246, 0.2)',
                              color: '#60a5fa',
                              fontSize: '0.8rem',
                              fontWeight: 600,
                              cursor: 'pointer'
                            }}
                          >
                            Transition
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

      {/* CREATE PARCEL MODAL */}
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '720px', maxHeight: '90vh', overflowY: 'auto', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Package size={22} color="var(--accent-cyan)" />
                Book & Register New Parcel
              </h3>
              <button onClick={() => setShowCreateModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            {createError && (
              <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--accent-rose)', color: 'var(--accent-rose)', marginBottom: '1rem', fontSize: '0.85rem' }}>
                {createError}
              </div>
            )}

            <form onSubmit={handleCreateSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1.2rem' }}>
              {/* Operating Company / Tenant Hub Network Selector */}
              <div style={{
                background: 'rgba(15, 23, 42, 0.6)',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '0.9rem',
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                  <label style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--accent-cyan)', display: 'flex', alignItems: 'center', gap: '0.4rem', margin: 0 }}>
                    <Building2 size={16} />
                    Operating Company & Tenant Hub Network *
                  </label>
                  {loadingTenantData && (
                    <span style={{ fontSize: '0.75rem', color: '#94a3b8', display: 'flex', alignItems: 'center', gap: '4px' }}>
                      <Loader2 size={12} className="spin-slow" /> Loading network hubs...
                    </span>
                  )}
                </div>
                <select
                  value={bookingTenantId}
                  onChange={(e) => handleBookingTenantChange(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.65rem 0.85rem',
                    borderRadius: '6px',
                    background: '#0f172a',
                    border: '1px solid #334155',
                    color: '#f8fafc',
                    fontSize: '0.9rem',
                    fontWeight: 600,
                    outline: 'none',
                  }}
                >
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name} ({t.slug}) — Role: {t.role}
                    </option>
                  ))}
                </select>
                <span style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: '6px', display: 'block' }}>
                  Select any company or tenant hub network you have authority over to dispatch parcels.
                </span>
              </div>

              {/* Routing Hubs */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Origin Branch (Intake Hub) *</label>
                  <select
                    required
                    value={formData.origin_branch_id}
                    onChange={(e) => setFormData({ ...formData, origin_branch_id: e.target.value })}
                    style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">Select Origin Hub</option>
                    {bookingBranches.map((b) => (
                      <option key={b.id} value={b.id}>{b.name} ({b.city})</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Destination Hub *</label>
                  <select
                    required
                    value={formData.destination_branch_id}
                    onChange={(e) => setFormData({ ...formData, destination_branch_id: e.target.value })}
                    style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">Select Destination Hub</option>
                    {bookingBranches.map((b) => (
                      <option key={b.id} value={b.id}>{b.name} ({b.city})</option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Service & Specs */}
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '0.75rem' }}>
                <div>
                  <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Service Level</label>
                  <select
                    value={formData.service_type}
                    onChange={(e) => setFormData({ ...formData, service_type: e.target.value as ServiceType })}
                    style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="STANDARD">Standard</option>
                    <option value="EXPRESS">Express (48h)</option>
                    <option value="OVERNIGHT">Overnight (24h)</option>
                    <option value="SAME_DAY">Same Day Priority</option>
                  </select>
                </div>

                <div>
                  <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Weight (kg) *</label>
                  <input
                    type="number"
                    step="0.1"
                    min="0.1"
                    required
                    value={formData.weight_kg}
                    onChange={(e) => setFormData({ ...formData, weight_kg: parseFloat(e.target.value) || 0 })}
                    style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                </div>

                <div>
                  <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Dimensions (cm)</label>
                  <input
                    type="text"
                    placeholder="30x20x15"
                    required
                    value={formData.dimensions_cm}
                    onChange={(e) => setFormData({ ...formData, dimensions_cm: e.target.value })}
                    style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                </div>

                <div>
                  <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Declared Value (₹)</label>
                  <input
                    type="number"
                    min="0"
                    step="50"
                    placeholder="0"
                    value={formData.declared_value || ''}
                    onChange={(e) => setFormData({ ...formData, declared_value: parseFloat(e.target.value) || 0 })}
                    style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                </div>
              </div>

              {/* Dynamic Price Estimate Card */}
              {(() => {
                const est = getEstimatedPrice(formData.service_type, formData.weight_kg, formData.declared_value);
                return (
                  <div style={{
                    background: 'rgba(6, 182, 212, 0.08)',
                    border: '1px solid rgba(6, 182, 212, 0.25)',
                    borderRadius: '8px',
                    padding: '0.85rem 1.1rem',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    flexWrap: 'wrap',
                    gap: '0.75rem',
                  }}>
                    <div>
                      <span style={{ fontSize: '0.8rem', color: '#94a3b8' }}>Dynamic Tariff & Insurance Estimate:</span>
                      <div style={{ fontSize: '1.25rem', fontWeight: 700, color: '#34d399' }}>₹{est.total.toFixed(2)}</div>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: '#cbd5e1', textAlign: 'right' }}>
                      Base: ₹{est.base} | Weight ({formData.weight_kg}kg): ₹{est.weightFee.toFixed(2)} | Insurance: ₹{est.insuranceFee.toFixed(2)}
                    </div>
                  </div>
                );
              })()}

              {/* Sender Details */}
              <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '1rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <h4 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--accent-cyan)', margin: 0 }}>Sender Information</h4>
                  {bookingCustomers.length > 0 && (
                    <select
                      value={formData.sender_customer_id || ''}
                      onChange={(e) => {
                        const custId = e.target.value;
                        const cust = bookingCustomers.find((c) => c.id === custId);
                        if (cust) {
                          setFormData((prev) => ({
                            ...prev,
                            sender_customer_id: cust.id,
                            sender_name: cust.name,
                            sender_phone: cust.phone,
                            sender_email: cust.email || '',
                            sender_address: cust.billing_address,
                          }));
                        } else {
                          setFormData((prev) => ({ ...prev, sender_customer_id: '' }));
                        }
                      }}
                      style={{
                        padding: '0.35rem 0.65rem',
                        borderRadius: '6px',
                        background: '#0f172a',
                        border: '1px solid #334155',
                        color: 'var(--accent-cyan)',
                        fontSize: '0.8rem',
                        outline: 'none',
                      }}
                    >
                      <option value="">⚡ Auto-fill from Customer CRM...</option>
                      {bookingCustomers.map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.customer_code} — {c.name} {c.company_name ? `(${c.company_name})` : ''}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', marginBottom: '0.75rem' }}>
                  <input
                    type="text"
                    placeholder="Sender Full Name *"
                    required
                    value={formData.sender_name}
                    onChange={(e) => setFormData({ ...formData, sender_name: e.target.value })}
                    style={{ padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                  <input
                    type="text"
                    placeholder="Sender Phone Number *"
                    required
                    value={formData.sender_phone}
                    onChange={(e) => setFormData({ ...formData, sender_phone: e.target.value })}
                    style={{ padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                </div>
                <input
                  type="text"
                  placeholder="Sender Complete Pickup/Origin Address *"
                  required
                  value={formData.sender_address}
                  onChange={(e) => setFormData({ ...formData, sender_address: e.target.value })}
                  style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              {/* Receiver Details */}
              <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '1rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <h4 style={{ fontSize: '0.95rem', fontWeight: 600, color: 'var(--accent-cyan)', margin: 0 }}>Recipient Information</h4>
                  {bookingCustomers.length > 0 && (
                    <select
                      value={formData.receiver_customer_id || ''}
                      onChange={(e) => {
                        const custId = e.target.value;
                        const cust = bookingCustomers.find((c) => c.id === custId);
                        if (cust) {
                          setFormData((prev) => ({
                            ...prev,
                            receiver_customer_id: cust.id,
                            receiver_name: cust.name,
                            receiver_phone: cust.phone,
                            receiver_email: cust.email || '',
                            receiver_address: cust.shipping_address || cust.billing_address,
                          }));
                        } else {
                          setFormData((prev) => ({ ...prev, receiver_customer_id: '' }));
                        }
                      }}
                      style={{
                        padding: '0.35rem 0.65rem',
                        borderRadius: '6px',
                        background: '#0f172a',
                        border: '1px solid #334155',
                        color: 'var(--accent-cyan)',
                        fontSize: '0.8rem',
                        outline: 'none',
                      }}
                    >
                      <option value="">⚡ Auto-fill from Customer CRM...</option>
                      {bookingCustomers.map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.customer_code} — {c.name} {c.company_name ? `(${c.company_name})` : ''}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem', marginBottom: '0.75rem' }}>
                  <input
                    type="text"
                    placeholder="Recipient Full Name *"
                    required
                    value={formData.receiver_name}
                    onChange={(e) => setFormData({ ...formData, receiver_name: e.target.value })}
                    style={{ padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                  <input
                    type="text"
                    placeholder="Recipient Phone Number *"
                    required
                    value={formData.receiver_phone}
                    onChange={(e) => setFormData({ ...formData, receiver_phone: e.target.value })}
                    style={{ padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  />
                </div>
                <input
                  type="text"
                  placeholder="Recipient Final Delivery Street Address *"
                  required
                  value={formData.receiver_address}
                  onChange={(e) => setFormData({ ...formData, receiver_address: e.target.value })}
                  style={{ width: '100%', padding: '0.6rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
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
                  disabled={createLoading}
                  style={{ padding: '0.6rem 1.4rem', borderRadius: '6px', border: 'none', background: 'var(--accent-cyan)', color: '#0f172a', fontWeight: 600, cursor: 'pointer' }}
                >
                  {createLoading ? 'Registering...' : 'Register Parcel & Generate QR'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* DETAIL & TIMELINE MODAL */}
      {showDetailModal && selectedParcel && (
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '780px', maxHeight: '90vh', overflowY: 'auto', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '1rem', flexWrap: 'wrap', gap: '0.75rem' }}>
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <span style={{ fontSize: '1.4rem', fontWeight: 700, fontFamily: 'monospace', color: 'var(--accent-cyan)' }}>
                    {selectedParcel.tracking_number}
                  </span>
                  <span style={{
                    padding: '2px 8px',
                    borderRadius: '12px',
                    fontSize: '0.75rem',
                    fontWeight: 700,
                    ...STATUS_COLORS[selectedParcel.status]
                  }}>
                    {selectedParcel.status}
                  </span>
                </div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '4px' }}>
                  Intake: {new Date(selectedParcel.created_at).toLocaleString()}
                </div>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <button
                  type="button"
                  onClick={() => {
                    setLabelParcel(selectedParcel);
                    setShowLabelModal(true);
                  }}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    padding: '0.55rem 1rem',
                    borderRadius: '8px',
                    border: '1px solid rgba(56, 189, 248, 0.4)',
                    background: 'rgba(56, 189, 248, 0.15)',
                    color: 'var(--accent-cyan)',
                    fontSize: '0.85rem',
                    fontWeight: 700,
                    cursor: 'pointer',
                  }}
                >
                  <Printer size={15} />
                  Print / Download Label & QR
                </button>

                <button onClick={() => setShowDetailModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', padding: '4px' }}>
                  <X size={20} />
                </button>
              </div>
            </div>

            {/* QR Code Payload display */}
            <div style={{
              background: 'rgba(15, 23, 42, 0.7)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '8px',
              padding: '1rem',
              marginBottom: '1.5rem',
              display: 'flex',
              gap: '1rem',
              alignItems: 'center',
            }}>
              <div style={{
                background: '#fff',
                padding: '12px',
                borderRadius: '8px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0
              }}>
                <QrCode size={64} color="#0f172a" />
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '4px' }}>
                  Tamper-Evident QR Security Signature
                </div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '6px' }}>
                  SHA-256 HMAC checksum embedded for contactless field barcode scanners.
                </div>
                <pre style={{
                  background: 'rgba(0, 0, 0, 0.4)',
                  padding: '6px 8px',
                  borderRadius: '4px',
                  fontSize: '0.7rem',
                  fontFamily: 'monospace',
                  color: 'var(--accent-cyan)',
                  overflowX: 'auto',
                  margin: 0
                }}>
                  {qrCodePayload || 'Generating QR payload...'}
                </pre>
              </div>
            </div>

            {/* Shipment details grid */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem', marginBottom: '1.5rem' }}>
              <div>
                <h4 style={{ fontSize: '0.85rem', textTransform: 'uppercase', color: 'var(--text-muted)', marginBottom: '8px' }}>Sender Origin</h4>
                <div style={{ fontSize: '0.9rem', fontWeight: 600 }}>{selectedParcel.sender_name}</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>{selectedParcel.sender_phone}</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '2px' }}>{selectedParcel.sender_address}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--accent-cyan)', marginTop: '4px' }}>Origin Hub: {selectedParcel.origin_branch_name}</div>
              </div>

              <div>
                <h4 style={{ fontSize: '0.85rem', textTransform: 'uppercase', color: 'var(--text-muted)', marginBottom: '8px' }}>Recipient Destination</h4>
                <div style={{ fontSize: '0.9rem', fontWeight: 600 }}>{selectedParcel.receiver_name}</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>{selectedParcel.receiver_phone}</div>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '2px' }}>{selectedParcel.receiver_address}</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--accent-emerald)', marginTop: '4px' }}>Destination Hub: {selectedParcel.destination_branch_name}</div>
              </div>
            </div>

            {/* Status Timeline Stepper */}
            <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '1.5rem' }}>
              <h4 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Clock size={18} color="var(--accent-cyan)" />
                Audit Trail & Milestone Timeline
              </h4>

              {timelineLoading ? (
                <div style={{ textAlign: 'center', padding: '1.5rem', color: 'var(--text-muted)' }}>Loading timeline events...</div>
              ) : timeline.length === 0 ? (
                <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>No status transitions recorded yet.</div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', position: 'relative', paddingLeft: '1.5rem' }}>
                  {timeline.map((evt, idx) => (
                    <div key={evt.id || idx} style={{ position: 'relative' }}>
                      <div style={{
                        position: 'absolute',
                        left: '-1.5rem',
                        top: '4px',
                        width: '10px',
                        height: '10px',
                        borderRadius: '50%',
                        background: 'var(--accent-cyan)',
                        boxShadow: '0 0 8px var(--accent-cyan)'
                      }} />
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                        <span style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--text-primary)' }}>
                          {evt.to_status.replace(/_/g, ' ')}
                        </span>
                        <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                          {new Date(evt.created_at).toLocaleString()}
                        </span>
                      </div>
                      {evt.branch_name && (
                        <div style={{ fontSize: '0.8rem', color: 'var(--accent-purple)' }}>
                          Location: {evt.branch_name}
                        </div>
                      )}
                      {evt.notes && (
                        <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '2px' }}>
                          {evt.notes}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* SCAN MODAL */}
      {showScanModal && (
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '560px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem' }}>
              <h3 style={{ fontSize: '1.25rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Scan size={22} color="var(--accent-purple)" />
                Field Barcode / QR Scanner
              </h3>
              <button onClick={() => setShowScanModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
              Paste raw QR scanner payload string or enter package tracking number to verify custody and recommended next actions.
            </p>

            <form onSubmit={handleScanSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <textarea
                rows={3}
                placeholder='e.g. {"chk":"...","parcel_id":"...","tenant_id":"...","tracking":"PKG-..."} or PKG-20260924-XXXX'
                value={scanInput}
                onChange={(e) => setScanInput(e.target.value)}
                style={{
                  width: '100%',
                  padding: '0.75rem',
                  borderRadius: '6px',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-subtle)',
                  color: 'var(--text-primary)',
                  fontFamily: 'monospace',
                  fontSize: '0.85rem',
                }}
              />

              <button
                type="submit"
                disabled={scanLoading || !scanInput.trim()}
                style={{
                  padding: '0.65rem 1.2rem',
                  borderRadius: '6px',
                  border: 'none',
                  background: 'var(--accent-purple)',
                  color: '#fff',
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
              >
                {scanLoading ? 'Verifying Code...' : 'Process Scanner Intake'}
              </button>
            </form>

            {scanError && (
              <div style={{ marginTop: '1rem', padding: '0.8rem', borderRadius: '6px', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--accent-rose)', color: 'var(--accent-rose)', fontSize: '0.85rem' }}>
                {scanError}
              </div>
            )}

            {scanResult && (
              <div style={{
                marginTop: '1.2rem',
                padding: '1.1rem',
                borderRadius: '8px',
                background: 'rgba(16, 185, 129, 0.12)',
                border: '1px solid rgba(16, 185, 129, 0.3)',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-emerald)', fontWeight: 700, fontSize: '0.95rem' }}>
                  <CheckCircle2 size={18} />
                  Verified Parcel Authenticity
                </div>
                <div style={{ marginTop: '0.5rem', fontSize: '0.85rem', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                  <div><strong>Tracking Code:</strong> <span style={{ fontFamily: 'monospace', color: 'var(--accent-cyan)' }}>{scanResult.parcel?.tracking_number}</span></div>
                  <div><strong>Parcel ID:</strong> <span style={{ fontFamily: 'monospace', fontSize: '0.78rem', color: 'var(--text-muted)' }}>{scanResult.parcel?.id}</span></div>
                  <div><strong>Current Status:</strong> <span style={{ color: '#38bdf8', fontWeight: 600 }}>{scanResult.parcel?.status}</span></div>
                  <div><strong>Recipient:</strong> {scanResult.parcel?.receiver_name} ({scanResult.parcel?.receiver_address})</div>
                  <div style={{ marginTop: '6px', padding: '6px 10px', borderRadius: '4px', background: 'rgba(255,255,255,0.06)', color: 'var(--accent-cyan)' }}>
                    <strong>Next Recommended Action:</strong> {scanResult.next_action}
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem', flexWrap: 'wrap' }}>
                  <button
                    type="button"
                    onClick={() => {
                      setLabelParcel(scanResult.parcel);
                      setShowLabelModal(true);
                      setShowScanModal(false);
                    }}
                    style={{
                      flex: 1,
                      minWidth: '160px',
                      padding: '0.55rem',
                      borderRadius: '6px',
                      border: '1px solid var(--border-subtle)',
                      background: 'rgba(30, 41, 59, 0.8)',
                      color: 'var(--text-primary)',
                      fontSize: '0.85rem',
                      fontWeight: 600,
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: '0.4rem',
                    }}
                  >
                    <QrCode size={15} color="var(--accent-cyan)" />
                    Label & QR Code
                  </button>

                  {(scanResult.parcel?.status === 'CREATED' || scanResult.parcel?.status === 'BOOKED') && (
                    <button
                      type="button"
                      onClick={async () => {
                        await handleQuickStatusAdvance(scanResult.parcel.id, 'RECEIVED_AT_ORIGIN_BRANCH', 'Intake via field scanner');
                        setShowScanModal(false);
                      }}
                      style={{
                        flex: 1,
                        minWidth: '160px',
                        padding: '0.55rem',
                        borderRadius: '6px',
                        border: 'none',
                        background: 'var(--accent-cyan)',
                        color: '#0f172a',
                        fontSize: '0.85rem',
                        fontWeight: 700,
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '0.4rem',
                      }}
                    >
                      <CheckCircle2 size={15} />
                      Confirm Hub Intake
                    </button>
                  )}

                  {scanResult.parcel?.status === 'IN_TRANSIT' && (
                    <button
                      type="button"
                      onClick={async () => {
                        await handleQuickStatusAdvance(scanResult.parcel.id, 'RECEIVED_AT_TRANSFER_BRANCH', 'Arrived at delivery hub via scanner');
                        setShowScanModal(false);
                      }}
                      style={{
                        flex: 1,
                        minWidth: '160px',
                        padding: '0.55rem',
                        borderRadius: '6px',
                        border: 'none',
                        background: 'var(--accent-emerald)',
                        color: '#0f172a',
                        fontSize: '0.85rem',
                        fontWeight: 700,
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '0.4rem',
                      }}
                    >
                      <CheckCircle2 size={15} />
                      Confirm Transfer Receipt
                    </button>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* QUICK STATUS UPDATE MODAL (Strict FSM Valid State Transitions) */}
      {showStatusModal && selectedParcel && (
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '480px', padding: '1.8rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem' }}>
              <div>
                <h3 style={{ fontSize: '1.2rem', fontWeight: 700, margin: 0 }}>
                  Advance Status: {selectedParcel.tracking_number}
                </h3>
                <span style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                  Current State: <strong style={{ color: 'var(--accent-cyan)' }}>{selectedParcel.status}</strong>
                </span>
              </div>
              <button onClick={() => setShowStatusModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            <form onSubmit={handleStatusSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                  Permitted Next State *
                </label>
                {(VALID_NEXT_TRANSITIONS[selectedParcel.status] || []).length === 0 ? (
                  <div style={{ padding: '0.75rem', borderRadius: '6px', background: 'rgba(239, 68, 68, 0.15)', color: '#ef4444', fontSize: '0.85rem' }}>
                    This parcel has reached terminal status ({selectedParcel.status}) and cannot be transitioned further.
                  </div>
                ) : (
                  <select
                    value={targetStatus}
                    onChange={(e) => setTargetStatus(e.target.value as ParcelStatus)}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    {(VALID_NEXT_TRANSITIONS[selectedParcel.status] || []).map((st) => (
                      <option key={st} value={st}>
                        {STATUS_LABELS[st] || st.replace(/_/g, ' ')}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div>
                <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Remarks / Handover Notes</label>
                <input
                  type="text"
                  placeholder="e.g. Scanned at Sorting Bay A"
                  value={statusNotes}
                  onChange={(e) => setStatusNotes(e.target.value)}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
                <button
                  type="button"
                  onClick={() => setShowStatusModal(false)}
                  style={{ padding: '0.6rem 1rem', borderRadius: '6px', border: '1px solid var(--border-subtle)', background: 'transparent', color: 'var(--text-secondary)', cursor: 'pointer' }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={statusLoading || (VALID_NEXT_TRANSITIONS[selectedParcel.status] || []).length === 0}
                  style={{ padding: '0.6rem 1.2rem', borderRadius: '6px', border: 'none', background: 'var(--accent-cyan)', color: '#0f172a', fontWeight: 600, cursor: 'pointer' }}
                >
                  {statusLoading ? 'Updating...' : 'Confirm Transition'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* SHIPPING LABEL & QR GENERATOR / DOWNLOAD MODAL */}
      <ParcelLabelModal
        parcel={labelParcel}
        branches={bookingBranches.length > 0 ? bookingBranches : branches}
        isOpen={showLabelModal}
        onClose={() => setShowLabelModal(false)}
        onStatusUpdate={async (parcelId, status) => {
          await handleQuickStatusAdvance(parcelId, status, 'Intake via Label Modal');
        }}
      />
    </div>
  );
};

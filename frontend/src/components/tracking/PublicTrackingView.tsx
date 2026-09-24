import React, { useState } from 'react';
import { 
  Search, 
  Package, 
  MapPin, 
  Clock, 
  CheckCircle2, 
  AlertCircle, 
  Shield, 
  Truck, 
  Sparkles
} from 'lucide-react';
import { api } from '../../services/api';
import type { PublicTrackingResponse, ParcelStatus } from '../../types/parcels';

const STAGES: { status: ParcelStatus[]; label: string; icon: any }[] = [
  { status: ['CREATED', 'BOOKED'], label: 'Order Booked', icon: Package },
  { status: ['RECEIVED_AT_ORIGIN_BRANCH'], label: 'Origin Hub Intake', icon: MapPin },
  { status: ['IN_TRANSIT', 'RECEIVED_AT_TRANSFER_BRANCH'], label: 'In Transit', icon: Truck },
  { status: ['OUT_FOR_DELIVERY', 'DELIVERY_ATTEMPTED'], label: 'Out for Delivery', icon: Clock },
  { status: ['DELIVERED'], label: 'Delivered', icon: CheckCircle2 },
];

export const PublicTrackingView: React.FC = () => {
  const [trackingNumber, setTrackingNumber] = useState('');
  const [trackingData, setTrackingData] = useState<PublicTrackingResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleTrack = async (e?: React.FormEvent, customNumber?: string) => {
    if (e) e.preventDefault();
    const query = (customNumber || trackingNumber).trim();
    if (!query) return;

    try {
      setLoading(true);
      setError(null);
      setTrackingData(null);
      const res = await api.getPublicTracking(query);
      setTrackingData(res);
    } catch (err: any) {
      setError(err.message || 'Tracking number not found. Please verify the code.');
    } finally {
      setLoading(false);
    }
  };

  const getCurrentStageIndex = (status: ParcelStatus): number => {
    if (status === 'DELIVERED') return 4;
    if (status === 'OUT_FOR_DELIVERY' || status === 'DELIVERY_ATTEMPTED') return 3;
    if (status === 'IN_TRANSIT' || status === 'RECEIVED_AT_TRANSFER_BRANCH') return 2;
    if (status === 'RECEIVED_AT_ORIGIN_BRANCH') return 1;
    return 0;
  };

  const stageIdx = trackingData ? getCurrentStageIndex(trackingData.status) : -1;

  return (
    <div style={{ maxWidth: '860px', margin: '0 auto', display: 'flex', flexDirection: 'column', gap: '2rem', paddingBottom: '3rem' }}>
      {/* Hero Header */}
      <div style={{ textAlign: 'center', padding: '2rem 1rem 1rem 1rem' }}>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', padding: '6px 14px', borderRadius: '20px', background: 'rgba(6, 182, 212, 0.12)', border: '1px solid rgba(6, 182, 212, 0.25)', color: 'var(--accent-cyan)', fontSize: '0.85rem', fontWeight: 600, marginBottom: '1rem' }}>
          <Sparkles size={15} />
          Customer Public Portal • Zero-PII Leak Architecture
        </div>
        <h1 style={{ fontSize: '2.4rem', fontWeight: 800, letterSpacing: '-0.02em', background: 'linear-gradient(135deg, #fff 0%, #94a3b8 100%)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
          Track Your Cargo Shipment
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1rem', maxWidth: '580px', margin: '0.75rem auto 0 auto' }}>
          Real-time milestones, verified route location telemetry, and scheduled arrival predictions.
        </p>

        {/* Search Bar */}
        <form onSubmit={handleTrack} style={{ marginTop: '2rem', display: 'flex', gap: '0.5rem', maxWidth: '560px', margin: '2rem auto 0 auto' }}>
          <div style={{ position: 'relative', flex: 1 }}>
            <Search size={18} style={{ position: 'absolute', left: '16px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }} />
            <input
              type="text"
              placeholder="Enter tracking number (e.g. PKG-20260924-...)"
              value={trackingNumber}
              onChange={(e) => setTrackingNumber(e.target.value)}
              style={{
                width: '100%',
                padding: '0.85rem 1rem 0.85rem 2.8rem',
                borderRadius: '10px',
                border: '1px solid var(--border-subtle)',
                background: 'rgba(15, 23, 42, 0.8)',
                color: 'var(--text-primary)',
                fontFamily: 'monospace',
                fontSize: '1rem',
                boxShadow: '0 4px 20px rgba(0, 0, 0, 0.4)',
              }}
            />
          </div>
          <button
            type="submit"
            disabled={loading || !trackingNumber.trim()}
            style={{
              padding: '0.85rem 1.8rem',
              borderRadius: '10px',
              border: 'none',
              background: 'linear-gradient(135deg, var(--accent-cyan) 0%, #2563eb 100%)',
              color: '#fff',
              fontWeight: 700,
              fontSize: '0.95rem',
              cursor: 'pointer',
              boxShadow: '0 4px 14px rgba(6, 182, 212, 0.3)',
            }}
          >
            {loading ? 'Locating...' : 'Track'}
          </button>
        </form>

        <div style={{ marginTop: '0.75rem', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
          <Shield size={12} style={{ display: 'inline', marginRight: '4px' }} />
          No login required. Sensitive personal identities and internal driver contact details are automatically masked.
        </div>
      </div>

      {error && (
        <div style={{
          padding: '1.2rem',
          borderRadius: '10px',
          background: 'rgba(244, 63, 94, 0.15)',
          border: '1px solid rgba(244, 63, 94, 0.3)',
          color: 'var(--accent-rose)',
          display: 'flex',
          alignItems: 'center',
          gap: '0.75rem',
          maxWidth: '560px',
          margin: '0 auto',
        }}>
          <AlertCircle size={22} style={{ flexShrink: 0 }} />
          <div style={{ fontSize: '0.9rem' }}>{error}</div>
        </div>
      )}

      {/* TRACKING RESULT DISPLAY */}
      {trackingData && (
        <div className="glass-panel" style={{ padding: '2rem', display: 'flex', flexDirection: 'column', gap: '2rem', border: '1px solid rgba(6, 182, 212, 0.3)' }}>
          {/* Top summary row */}
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '1rem', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '1.5rem' }}>
            <div>
              <div style={{ fontSize: '0.8rem', textTransform: 'uppercase', color: 'var(--text-muted)', fontWeight: 600 }}>Tracking Identifier</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 800, fontFamily: 'monospace', color: 'var(--accent-cyan)', marginTop: '2px' }}>
                {trackingData.tracking_number}
              </div>
              <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '4px' }}>
                Service Level: <strong>{trackingData.service_type}</strong>
              </div>
            </div>

            <div style={{ textAlign: 'right' }}>
              <span style={{
                display: 'inline-block',
                padding: '6px 16px',
                borderRadius: '20px',
                fontSize: '0.85rem',
                fontWeight: 700,
                background: trackingData.status === 'DELIVERED' ? 'rgba(16, 185, 129, 0.2)' : 'rgba(6, 182, 212, 0.2)',
                color: trackingData.status === 'DELIVERED' ? 'var(--accent-emerald)' : 'var(--accent-cyan)',
                border: '1px solid currentColor',
              }}>
                {trackingData.status_display}
              </span>
              <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: '6px' }}>
                Last updated: {new Date(trackingData.updated_at).toLocaleTimeString()}
              </div>
            </div>
          </div>

          {/* Route Info */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-around', background: 'rgba(15, 23, 42, 0.6)', padding: '1.2rem', borderRadius: '10px' }}>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Origin</div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-primary)', marginTop: '2px' }}>{trackingData.origin_city}</div>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-cyan)' }}>
              <div style={{ width: '40px', height: '2px', background: 'var(--border-subtle)' }} />
              <Truck size={20} />
              <div style={{ width: '40px', height: '2px', background: 'var(--border-subtle)' }} />
            </div>

            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Destination</div>
              <div style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-primary)', marginTop: '2px' }}>{trackingData.destination_city}</div>
            </div>
          </div>

          {/* Visual Milestone Stepper */}
          <div>
            <h3 style={{ fontSize: '1rem', fontWeight: 700, marginBottom: '1.2rem' }}>Delivery Progress</h3>
            <div style={{ display: 'flex', justifyContent: 'space-between', position: 'relative' }}>
              {/* Connecting line */}
              <div style={{
                position: 'absolute',
                top: '18px',
                left: '20px',
                right: '20px',
                height: '3px',
                background: 'rgba(255, 255, 255, 0.1)',
                zIndex: 1,
              }} />
              <div style={{
                position: 'absolute',
                top: '18px',
                left: '20px',
                width: `${(stageIdx / (STAGES.length - 1)) * 90}%`,
                height: '3px',
                background: 'linear-gradient(90deg, var(--accent-cyan), var(--accent-emerald))',
                zIndex: 2,
                transition: 'width 0.4s ease',
              }} />

              {STAGES.map((st, i) => {
                const isPassed = i <= stageIdx;
                const isCurrent = i === stageIdx;
                const IconComponent = st.icon;
                return (
                  <div key={st.label} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', zIndex: 3, width: '100px', textAlign: 'center' }}>
                    <div style={{
                      width: '38px',
                      height: '38px',
                      borderRadius: '50%',
                      background: isCurrent ? 'var(--accent-cyan)' : isPassed ? 'var(--accent-emerald)' : 'rgba(30, 41, 59, 0.9)',
                      color: isCurrent || isPassed ? '#0f172a' : 'var(--text-muted)',
                      border: isCurrent ? '3px solid #fff' : '2px solid rgba(255, 255, 255, 0.15)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      boxShadow: isCurrent ? '0 0 14px var(--accent-cyan)' : 'none',
                    }}>
                      <IconComponent size={18} />
                    </div>
                    <span style={{
                      fontSize: '0.75rem',
                      fontWeight: isCurrent ? 700 : 500,
                      color: isCurrent ? 'var(--accent-cyan)' : isPassed ? 'var(--text-primary)' : 'var(--text-muted)',
                      marginTop: '8px',
                    }}>
                      {st.label}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Milestone Audit Trail */}
          <div style={{ borderTop: '1px solid var(--border-subtle)', paddingTop: '1.5rem' }}>
            <h3 style={{ fontSize: '1rem', fontWeight: 700, marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Clock size={16} color="var(--accent-cyan)" />
              Shipment Activity History
            </h3>

            {(!trackingData.milestones || trackingData.milestones.length === 0) ? (
              <div style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>No activity milestones recorded yet.</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.8rem', paddingLeft: '1rem', borderLeft: '2px solid var(--border-subtle)' }}>
                {trackingData.milestones.map((m, idx) => (
                  <div key={idx} style={{ position: 'relative', paddingLeft: '0.5rem' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                      <span style={{ fontWeight: 600, fontSize: '0.9rem', color: 'var(--text-primary)' }}>
                        {m.description}
                      </span>
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)', fontFamily: 'monospace' }}>
                        {new Date(m.timestamp).toLocaleString()}
                      </span>
                    </div>
                    {m.location && (
                      <div style={{ fontSize: '0.8rem', color: 'var(--accent-purple)', marginTop: '2px' }}>
                        <MapPin size={11} style={{ display: 'inline', marginRight: '3px' }} />
                        {m.location}
                      </div>
                    )}
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

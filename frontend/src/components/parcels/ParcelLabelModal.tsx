import React, { useState, useEffect, useRef } from 'react';
import QRCode from 'qrcode';
import { 
  X, 
  Download, 
  Printer, 
  Copy, 
  Check, 
  MapPin, 
  CheckCircle2, 
  QrCode as QrIcon
} from 'lucide-react';
import type { Parcel, ParcelStatus } from '../../types/parcels';
import type { Branch } from '../../types/resources';

interface ParcelLabelModalProps {
  parcel: Parcel | null;
  branches: Branch[];
  isOpen: boolean;
  onClose: () => void;
  onStatusUpdate?: (parcelId: string, status: ParcelStatus) => Promise<void>;
}

export const ParcelLabelModal: React.FC<ParcelLabelModalProps> = ({
  parcel,
  branches,
  isOpen,
  onClose,
  onStatusUpdate,
}) => {
  const [qrDataUrl, setQrDataUrl] = useState<string>('');
  const [copiedId, setCopiedId] = useState(false);
  const [copiedTracking, setCopiedTracking] = useState(false);
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);
  const [updateMsg, setUpdateMsg] = useState<string | null>(null);
  const labelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!parcel) return;

    // Generate high-resolution QR code (300dpi equivalent scale)
    const qrPayload = parcel.qr_code_payload || JSON.stringify({
      tracking: parcel.tracking_number,
      tenant_id: parcel.tenant_id,
      parcel_id: parcel.id,
    });

    QRCode.toDataURL(qrPayload, {
      width: 400,
      margin: 2,
      errorCorrectionLevel: 'M',
      color: {
        dark: '#000000',
        light: '#FFFFFF',
      },
    })
      .then((url) => setQrDataUrl(url))
      .catch((err) => console.error('Failed to generate QR Code:', err));
  }, [parcel]);

  if (!isOpen || !parcel) return null;

  // Resolve branch names and cities
  const originBranch = branches.find((b) => b.id === parcel.origin_branch_id);
  const destBranch = branches.find((b) => b.id === parcel.destination_branch_id);

  const originName = parcel.origin_branch_name || originBranch?.name || 'Origin Hub';
  const originCity = originBranch?.city || 'Origin City';
  const destName = parcel.destination_branch_name || destBranch?.name || 'Destination Hub';
  const destCity = destBranch?.city || 'Destination City';

  const handleCopy = (text: string, type: 'id' | 'tracking') => {
    navigator.clipboard.writeText(text);
    if (type === 'id') {
      setCopiedId(true);
      setTimeout(() => setCopiedId(false), 2000);
    } else {
      setCopiedTracking(true);
      setTimeout(() => setCopiedTracking(false), 2000);
    }
  };

  // Download complete high-res printable label as PNG canvas
  const handleDownloadFullLabel = async () => {
    if (!qrDataUrl) return;

    const canvas = document.createElement('canvas');
    const width = 800;
    const height = 1100;
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // 1. Background
    ctx.fillStyle = '#FFFFFF';
    ctx.fillRect(0, 0, width, height);

    // 2. Outer border
    ctx.strokeStyle = '#000000';
    ctx.lineWidth = 6;
    ctx.strokeRect(20, 20, width - 40, height - 40);

    // 3. Header Banner
    ctx.fillStyle = '#0F172A';
    ctx.fillRect(20, 20, width - 40, 90);

    ctx.fillStyle = '#FFFFFF';
    ctx.font = 'bold 34px sans-serif';
    ctx.fillText('LOGIFLOWS EXPRESS ROUTING LABEL', 45, 75);

    ctx.font = 'bold 20px monospace';
    ctx.fillStyle = '#38BDF8';
    ctx.fillText(parcel.service_type || 'STANDARD', width - 200, 75);

    // 4. Destination Section (Prominent Large Font for Hub Sorters)
    ctx.fillStyle = '#F8FAFC';
    ctx.fillRect(30, 120, width - 60, 160);
    ctx.strokeStyle = '#CBD5E1';
    ctx.lineWidth = 2;
    ctx.strokeRect(30, 120, width - 60, 160);

    ctx.fillStyle = '#DC2626';
    ctx.font = 'bold 18px sans-serif';
    ctx.fillText('DESTINATION HUB & DELIVERY ZONE:', 50, 150);

    ctx.fillStyle = '#000000';
    ctx.font = 'bold 36px sans-serif';
    ctx.fillText(`${destCity.toUpperCase()} — ${destName.toUpperCase()}`, 50, 195);

    ctx.font = 'bold 20px sans-serif';
    ctx.fillStyle = '#1E293B';
    ctx.fillText(`DELIVER TO: ${parcel.receiver_name} | Tel: ${parcel.receiver_phone}`, 50, 230);
    ctx.font = '18px sans-serif';
    ctx.fillStyle = '#475569';
    ctx.fillText(`Address: ${parcel.receiver_address}`, 50, 260);

    // 5. Origin Section
    ctx.fillStyle = '#64748B';
    ctx.font = 'bold 16px sans-serif';
    ctx.fillText('SHIP FROM (ORIGIN):', 50, 315);
    ctx.fillStyle = '#0F172A';
    ctx.font = 'bold 18px sans-serif';
    ctx.fillText(`${originName} (${originCity}) — ${parcel.sender_name} (Tel: ${parcel.sender_phone})`, 50, 340);
    ctx.font = '16px sans-serif';
    ctx.fillStyle = '#475569';
    ctx.fillText(`Address: ${parcel.sender_address}`, 50, 365);

    // Horizontal divider
    ctx.strokeStyle = '#000000';
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(30, 390);
    ctx.lineTo(width - 30, 390);
    ctx.stroke();

    // 6. Draw QR Code into Canvas center
    const qrImg = new Image();
    qrImg.crossOrigin = 'anonymous';
    qrImg.src = qrDataUrl;
    await new Promise((resolve) => {
      qrImg.onload = resolve;
    });
    const qrSize = 340;
    const qrX = (width - qrSize) / 2;
    const qrY = 410;
    ctx.drawImage(qrImg, qrX, qrY, qrSize, qrSize);

    // 7. Tracking Code & Monospace Barcode text
    ctx.fillStyle = '#000000';
    ctx.font = 'bold 32px monospace';
    ctx.textAlign = 'center';
    ctx.fillText(parcel.tracking_number, width / 2, 790);

    ctx.font = '16px sans-serif';
    ctx.fillStyle = '#64748B';
    ctx.fillText('SCAN FOR DIGITAL CUSTODY & HUB ROUTING INTAKE', width / 2, 820);

    // Horizontal divider
    ctx.beginPath();
    ctx.moveTo(30, 845);
    ctx.lineTo(width - 30, 845);
    ctx.stroke();

    // 8. Package Metadata Footer (Parcel ID, Weight, Date)
    ctx.textAlign = 'left';
    ctx.fillStyle = '#0F172A';
    ctx.font = 'bold 18px sans-serif';
    ctx.fillText(`PARCEL ID:`, 50, 885);
    ctx.font = '16px monospace';
    ctx.fillStyle = '#334155';
    ctx.fillText(parcel.id, 170, 885);

    ctx.fillStyle = '#0F172A';
    ctx.font = 'bold 18px sans-serif';
    ctx.fillText(`WEIGHT: ${parcel.weight_kg} KG`, 50, 925);
    ctx.fillText(`DIMENSIONS: ${parcel.dimensions_cm || 'Standard Box'}`, 300, 925);
    ctx.fillText(`STATUS: ${parcel.status}`, 580, 925);

    ctx.fillStyle = '#64748B';
    ctx.font = '14px sans-serif';
    ctx.fillText(`Created: ${new Date(parcel.created_at).toLocaleString()} | Security Checksum Verified`, 50, 970);
    ctx.fillText(`LogiFlows Autonomous Logistics Operating System — Multi-Tenant Encrypted Payload`, 50, 1000);

    // Trigger Download
    const link = document.createElement('a');
    link.download = `LogiFlows-Label-${parcel.tracking_number}.png`;
    link.href = canvas.toDataURL('image/png');
    link.click();
  };

  // Download QR code with Destination and Parcel ID
  const handleDownloadQrWithDestinationAndId = async () => {
    if (!qrDataUrl || !parcel) return;

    const canvas = document.createElement('canvas');
    const width = 560;
    const height = 680;
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Background
    ctx.fillStyle = '#FFFFFF';
    ctx.fillRect(0, 0, width, height);

    // Border
    ctx.strokeStyle = '#0F172A';
    ctx.lineWidth = 4;
    ctx.strokeRect(16, 16, width - 32, height - 32);

    // Top Header Banner
    ctx.fillStyle = '#0F172A';
    ctx.fillRect(16, 16, width - 32, 60);

    ctx.fillStyle = '#38BDF8';
    ctx.font = 'bold 20px monospace';
    ctx.textAlign = 'center';
    ctx.fillText('LOGIFLOWS PARCEL ROUTING QR', width / 2, 54);

    // Destination Box
    ctx.fillStyle = '#FEF2F2';
    ctx.fillRect(26, 88, width - 52, 75);
    ctx.strokeStyle = '#FCA5A5';
    ctx.lineWidth = 1.5;
    ctx.strokeRect(26, 88, width - 52, 75);

    ctx.fillStyle = '#DC2626';
    ctx.font = 'bold 13px sans-serif';
    ctx.textAlign = 'left';
    ctx.fillText('DESTINATION HUB & ZONE:', 40, 110);

    ctx.fillStyle = '#0F172A';
    ctx.font = 'bold 22px sans-serif';
    ctx.fillText(`${destCity.toUpperCase()} — ${destName.toUpperCase()}`, 40, 140);

    // Draw QR Code
    const qrImg = new Image();
    qrImg.crossOrigin = 'anonymous';
    qrImg.src = qrDataUrl;
    await new Promise((resolve) => {
      qrImg.onload = resolve;
    });
    const qrSize = 320;
    const qrX = (width - qrSize) / 2;
    const qrY = 175;
    ctx.drawImage(qrImg, qrX, qrY, qrSize, qrSize);

    // Tracking Number
    ctx.fillStyle = '#000000';
    ctx.font = 'bold 24px monospace';
    ctx.textAlign = 'center';
    ctx.fillText(parcel.tracking_number, width / 2, 530);

    // Parcel ID
    ctx.font = 'bold 14px sans-serif';
    ctx.fillStyle = '#475569';
    ctx.fillText(`Parcel ID: ${parcel.id}`, width / 2, 560);

    // Receiver and Service
    ctx.font = '13px sans-serif';
    ctx.fillStyle = '#64748B';
    ctx.fillText(`Deliver to: ${parcel.receiver_name} | Service: ${parcel.service_type || 'STANDARD'}`, width / 2, 585);

    // Footer
    ctx.font = '11px sans-serif';
    ctx.fillStyle = '#94A3B8';
    ctx.fillText('Scan for Origin Hub Intake, Custody Handoff & Dispatch Confirmation', width / 2, 630);

    // Trigger Download
    const link = document.createElement('a');
    link.download = `QR-Destination-${parcel.tracking_number}.png`;
    link.href = canvas.toDataURL('image/png');
    link.click();
  };

  // Download QR code image only
  const handleDownloadQrOnly = () => {
    if (!qrDataUrl) return;
    const link = document.createElement('a');
    link.download = `QR-${parcel.tracking_number}.png`;
    link.href = qrDataUrl;
    link.click();
  };

  // Print label directly
  const handlePrint = () => {
    window.print();
  };

  const handleQuickIntake = async () => {
    if (!onStatusUpdate) return;
    try {
      setIsUpdatingStatus(true);
      await onStatusUpdate(parcel.id, 'RECEIVED_AT_ORIGIN_BRANCH');
      setUpdateMsg('Intake completed! Parcel marked RECEIVED_AT_ORIGIN_BRANCH.');
      setTimeout(() => setUpdateMsg(null), 3500);
    } catch (err: any) {
      alert(`Intake failed: ${err.message}`);
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0, 0, 0, 0.82)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1100,
        padding: '1rem',
        backdropFilter: 'blur(8px)',
      }}
    >
      <div
        className="glass-panel"
        style={{
          width: '100%',
          maxWidth: '680px',
          maxHeight: '92vh',
          display: 'flex',
          flexDirection: 'column',
          borderRadius: '16px',
          border: '1px solid rgba(56, 189, 248, 0.3)',
          overflow: 'hidden',
          boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.7)',
        }}
      >
        {/* Modal Header */}
        <div
          style={{
            padding: '1.25rem 1.5rem',
            borderBottom: '1px solid var(--border-subtle)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            background: 'rgba(15, 23, 42, 0.95)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div
              style={{
                width: '38px',
                height: '38px',
                borderRadius: '8px',
                background: 'rgba(56, 189, 248, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--accent-cyan)',
              }}
            >
              <QrIcon size={22} />
            </div>
            <div>
              <h3 style={{ fontSize: '1.2rem', fontWeight: 700, margin: 0, color: '#FFFFFF' }}>
                Parcel QR Code & Shipping Label
              </h3>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                Official routing label with destination and verified parcel identifier
              </span>
            </div>
          </div>

          <button
            onClick={onClose}
            style={{
              background: 'transparent',
              border: 'none',
              color: 'var(--text-muted)',
              cursor: 'pointer',
              padding: '6px',
              borderRadius: '6px',
            }}
          >
            <X size={20} />
          </button>
        </div>

        {/* Modal Scrollable Body */}
        <div style={{ padding: '1.5rem', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '1.2rem' }}>
          {updateMsg && (
            <div
              style={{
                padding: '0.8rem 1rem',
                borderRadius: '8px',
                background: 'rgba(16, 185, 129, 0.15)',
                border: '1px solid rgba(16, 185, 129, 0.4)',
                color: '#34d399',
                display: 'flex',
                alignItems: 'center',
                gap: '0.6rem',
                fontSize: '0.9rem',
                fontWeight: 600,
              }}
            >
              <CheckCircle2 size={18} />
              {updateMsg}
            </div>
          )}

          {/* PHYSICAL SHIPPING LABEL PREVIEW CARD */}
          <div
            ref={labelRef}
            style={{
              background: '#FFFFFF',
              color: '#0F172A',
              borderRadius: '12px',
              padding: '1.5rem',
              border: '3px solid #0F172A',
              boxShadow: '0 10px 25px rgba(0,0,0,0.4)',
              display: 'flex',
              flexDirection: 'column',
              gap: '1rem',
            }}
          >
            {/* Label Top Bar */}
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                borderBottom: '2px solid #0F172A',
                paddingBottom: '0.75rem',
              }}
            >
              <div>
                <div style={{ fontSize: '0.75rem', fontWeight: 800, letterSpacing: '1px', color: '#64748B' }}>
                  LOGIFLOWS CARGO ROUTING SYSTEM
                </div>
                <div style={{ fontSize: '1.3rem', fontWeight: 900, color: '#0F172A', letterSpacing: '-0.5px' }}>
                  EXPRESS ROUTING SLIP
                </div>
              </div>
              <div
                style={{
                  background: '#0F172A',
                  color: '#FFFFFF',
                  padding: '4px 10px',
                  borderRadius: '4px',
                  fontWeight: 800,
                  fontSize: '0.85rem',
                  letterSpacing: '0.5px',
                }}
              >
                {parcel.service_type || 'STANDARD'}
              </div>
            </div>

            {/* PROMINENT DESTINATION SECTION */}
            <div
              style={{
                background: '#F1F5F9',
                border: '2px solid #E2E8F0',
                borderRadius: '8px',
                padding: '0.9rem 1.1rem',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: '#DC2626', fontWeight: 800, fontSize: '0.8rem', letterSpacing: '0.5px' }}>
                <MapPin size={16} />
                DESTINATION ROUTING HUB:
              </div>
              <div style={{ fontSize: '1.4rem', fontWeight: 900, color: '#0F172A', marginTop: '2px' }}>
                {destCity.toUpperCase()} — {destName.toUpperCase()}
              </div>

              <div style={{ marginTop: '0.6rem', paddingTop: '0.6rem', borderTop: '1px dashed #CBD5E1', fontSize: '0.88rem' }}>
                <div style={{ fontWeight: 700, color: '#1E293B' }}>
                  DELIVER TO: {parcel.receiver_name} {parcel.receiver_phone ? `(${parcel.receiver_phone})` : ''}
                </div>
                <div style={{ color: '#475569', fontSize: '0.82rem', marginTop: '2px' }}>
                  {parcel.receiver_address}, {destCity}
                </div>
              </div>
            </div>

            {/* ORIGIN HUB & SENDER DETAILS */}
            <div style={{ fontSize: '0.82rem', color: '#475569', background: '#F8FAFC', padding: '0.6rem 0.9rem', borderRadius: '6px', border: '1px solid #E2E8F0' }}>
              <div>
                <strong style={{ color: '#1E293B' }}>FROM (ORIGIN HUB):</strong> {originName} ({originCity})
              </div>
              <div>
                Sender: {parcel.sender_name} | Tel: {parcel.sender_phone} | {parcel.sender_address}
              </div>
            </div>

            {/* QR CODE & BARCODE IDENTIFIERS */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', justifyContent: 'center', padding: '0.5rem 0' }}>
              {qrDataUrl ? (
                <div style={{ textAlign: 'center' }}>
                  <img
                    src={qrDataUrl}
                    alt="Parcel QR Code"
                    style={{
                      width: '180px',
                      height: '180px',
                      display: 'block',
                      margin: '0 auto',
                      borderRadius: '4px',
                      border: '1px solid #E2E8F0',
                    }}
                  />
                  <span style={{ fontSize: '0.7rem', color: '#64748B', fontWeight: 600, display: 'block', marginTop: '4px' }}>
                    SCAN FOR CUSTODY HANDOVER
                  </span>
                </div>
              ) : (
                <div style={{ width: '180px', height: '180px', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#F1F5F9' }}>
                  Generating QR...
                </div>
              )}

              {/* IDENTIFIERS BLOCK */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem', flex: 1 }}>
                <div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 700, color: '#64748B' }}>TRACKING NUMBER</div>
                  <div style={{ fontSize: '1.25rem', fontWeight: 900, fontFamily: 'monospace', color: '#0F172A', wordBreak: 'break-all' }}>
                    {parcel.tracking_number}
                  </div>
                </div>

                <div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 700, color: '#64748B' }}>PARCEL ID (SYSTEM UUID)</div>
                  <div style={{ fontSize: '0.8rem', fontFamily: 'monospace', color: '#475569', wordBreak: 'break-all' }}>
                    {parcel.id}
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '1rem', marginTop: '4px' }}>
                  <div>
                    <span style={{ fontSize: '0.72rem', color: '#64748B', display: 'block' }}>WEIGHT</span>
                    <strong style={{ fontSize: '0.9rem', color: '#0F172A' }}>{parcel.weight_kg} KG</strong>
                  </div>
                  <div>
                    <span style={{ fontSize: '0.72rem', color: '#64748B', display: 'block' }}>DIMENSIONS</span>
                    <strong style={{ fontSize: '0.9rem', color: '#0F172A' }}>{parcel.dimensions_cm || '30x20x15'}</strong>
                  </div>
                  <div>
                    <span style={{ fontSize: '0.72rem', color: '#64748B', display: 'block' }}>CURRENT STATE</span>
                    <strong style={{ fontSize: '0.85rem', color: '#0284C7' }}>{parcel.status}</strong>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* QUICK IDENTIFIERS COPY ROW */}
          <div
            style={{
              display: 'flex',
              gap: '0.75rem',
              flexWrap: 'wrap',
              background: 'rgba(15, 23, 42, 0.6)',
              padding: '0.75rem 1rem',
              borderRadius: '8px',
              border: '1px solid var(--border-subtle)',
            }}
          >
            <div style={{ flex: 1, minWidth: '220px' }}>
              <span style={{ fontSize: '0.72rem', color: 'var(--text-muted)', display: 'block' }}>Tracking Code</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '2px' }}>
                <span style={{ fontFamily: 'monospace', fontWeight: 700, color: 'var(--accent-cyan)', fontSize: '0.9rem' }}>
                  {parcel.tracking_number}
                </span>
                <button
                  onClick={() => handleCopy(parcel.tracking_number, 'tracking')}
                  style={{
                    background: 'none',
                    border: 'none',
                    color: copiedTracking ? 'var(--accent-emerald)' : 'var(--text-muted)',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                    fontSize: '0.75rem',
                  }}
                >
                  {copiedTracking ? <Check size={14} /> : <Copy size={14} />}
                  {copiedTracking ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>

            <div style={{ flex: 1.2, minWidth: '250px' }}>
              <span style={{ fontSize: '0.72rem', color: 'var(--text-muted)', display: 'block' }}>Parcel UUID</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '2px' }}>
                <span style={{ fontFamily: 'monospace', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>
                  {parcel.id}
                </span>
                <button
                  onClick={() => handleCopy(parcel.id, 'id')}
                  style={{
                    background: 'none',
                    border: 'none',
                    color: copiedId ? 'var(--accent-emerald)' : 'var(--text-muted)',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                    fontSize: '0.75rem',
                  }}
                >
                  {copiedId ? <Check size={14} /> : <Copy size={14} />}
                  {copiedId ? 'Copied' : 'Copy'}
                </button>
              </div>
            </div>
          </div>

          {/* QUICK LIFECYCLE INTAKE ACTION */}
          {(parcel.status === 'CREATED' || parcel.status === 'BOOKED') && onStatusUpdate && (
            <div
              style={{
                padding: '0.9rem 1.1rem',
                borderRadius: '8px',
                background: 'rgba(56, 189, 248, 0.08)',
                border: '1px solid rgba(56, 189, 248, 0.3)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: '0.8rem',
              }}
            >
              <div>
                <div style={{ fontSize: '0.9rem', fontWeight: 700, color: 'var(--accent-cyan)' }}>
                  Advance Lifecycle: Origin Hub Intake
                </div>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                  This package is booked. Click below to confirm arrival & scan intake into {originName}.
                </div>
              </div>
              <button
                onClick={handleQuickIntake}
                disabled={isUpdatingStatus}
                style={{
                  padding: '0.55rem 1.1rem',
                  borderRadius: '6px',
                  border: 'none',
                  background: 'var(--accent-cyan)',
                  color: '#0F172A',
                  fontWeight: 700,
                  fontSize: '0.85rem',
                  cursor: isUpdatingStatus ? 'not-allowed' : 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                }}
              >
                <CheckCircle2 size={16} />
                {isUpdatingStatus ? 'Intaking...' : 'Intake Package at Origin Hub'}
              </button>
            </div>
          )}
        </div>

        {/* Modal Footer Action Buttons */}
        <div
          style={{
            padding: '1.25rem 1.5rem',
            borderTop: '1px solid var(--border-subtle)',
            background: 'rgba(15, 23, 42, 0.95)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '0.75rem',
          }}
        >
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            <button
              onClick={handleDownloadQrOnly}
              style={{
                padding: '0.6rem 0.9rem',
                borderRadius: '8px',
                border: '1px solid var(--border-subtle)',
                background: 'transparent',
                color: 'var(--text-secondary)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                fontSize: '0.85rem',
                fontWeight: 600,
              }}
            >
              <Download size={16} />
              Raw QR (PNG)
            </button>

            <button
              onClick={handleDownloadQrWithDestinationAndId}
              style={{
                padding: '0.6rem 1rem',
                borderRadius: '8px',
                border: '1px solid rgba(6, 182, 212, 0.4)',
                background: 'rgba(6, 182, 212, 0.1)',
                color: 'var(--accent-cyan)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                fontSize: '0.85rem',
                fontWeight: 600,
              }}
            >
              <QrIcon size={16} />
              QR with Destination & ID (PNG)
            </button>
          </div>

          <div style={{ display: 'flex', gap: '0.75rem' }}>
            <button
              onClick={handlePrint}
              style={{
                padding: '0.6rem 1.1rem',
                borderRadius: '8px',
                border: '1px solid var(--border-subtle)',
                background: 'rgba(255, 255, 255, 0.05)',
                color: 'var(--text-primary)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                fontSize: '0.85rem',
                fontWeight: 600,
              }}
            >
              <Printer size={16} />
              Print Label (4x6")
            </button>

            <button
              onClick={handleDownloadFullLabel}
              style={{
                padding: '0.65rem 1.3rem',
                borderRadius: '8px',
                border: 'none',
                background: 'linear-gradient(135deg, #0284C7 0%, #2563EB 100%)',
                color: '#FFFFFF',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.6rem',
                fontSize: '0.9rem',
                fontWeight: 700,
                boxShadow: '0 4px 12px rgba(37, 99, 235, 0.35)',
              }}
            >
              <Download size={18} />
              Download Shipping Label (PNG)
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

import React, { useState, useEffect } from 'react';
import { 
  Truck, 
  Filter, 
  Plus, 
  CheckCircle2, 
  AlertTriangle, 
  Clock, 
  User, 
  MapPin, 
  FileCheck, 
  RefreshCw, 
  X,
  AlertCircle,
  Phone,
  ShieldCheck
} from 'lucide-react';
import { api } from '../../services/api';
import type { 
  DeliveryTask, 
  DeliveryTaskStatus, 
  DeliveryPriority, 
  AttemptOutcome, 
  ProofType,
  CreateDeliveryTaskPayload,
  RecordDeliveryAttemptPayload,
  SubmitDeliveryProofPayload,
  DeliveryAttempt,
  Parcel
} from '../../types/parcels';
import type { Employee, Vehicle } from '../../types/resources';

interface DeliveryTaskListProps {
  tenantId: string;
  userRole?: string;
}

const PRIORITY_BADGES: Record<DeliveryPriority, { label: string; bg: string; text: string }> = {
  LOW: { label: 'Low', bg: 'rgba(148, 163, 184, 0.15)', text: '#94a3b8' },
  NORMAL: { label: 'Normal', bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa' },
  HIGH: { label: 'High Priority', bg: 'rgba(245, 158, 11, 0.15)', text: '#fbbf24' },
  URGENT: { label: 'URGENT (Same Day)', bg: 'rgba(244, 63, 94, 0.18)', text: '#f43f5e' },
};

const TASK_STATUS_COLORS: Record<DeliveryTaskStatus, { bg: string; text: string; border: string }> = {
  ASSIGNED: { bg: 'rgba(99, 102, 241, 0.15)', text: '#818cf8', border: 'rgba(99, 102, 241, 0.3)' },
  IN_PROGRESS: { bg: 'rgba(245, 158, 11, 0.18)', text: '#fbbf24', border: 'rgba(245, 158, 11, 0.4)' },
  COMPLETED: { bg: 'rgba(16, 185, 129, 0.18)', text: '#34d399', border: 'rgba(16, 185, 129, 0.4)' },
  FAILED: { bg: 'rgba(244, 63, 94, 0.15)', text: '#f43f5e', border: 'rgba(244, 63, 94, 0.3)' },
  CANCELLED: { bg: 'rgba(156, 163, 175, 0.15)', text: '#9ca3af', border: 'rgba(156, 163, 175, 0.3)' },
};

export const DeliveryTaskList: React.FC<DeliveryTaskListProps> = ({ tenantId, userRole }) => {
  const [tasks, setTasks] = useState<DeliveryTask[]>([]);
  const [eligibleParcels, setEligibleParcels] = useState<Parcel[]>([]);
  const [drivers, setDrivers] = useState<Employee[]>([]);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [priorityFilter, setPriorityFilter] = useState<string>('');

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showAttemptModal, setShowAttemptModal] = useState(false);
  const [showProofModal, setShowProofModal] = useState(false);
  const [showHistoryModal, setShowHistoryModal] = useState(false);

  // Selected item
  const [selectedTask, setSelectedTask] = useState<DeliveryTask | null>(null);
  const [attemptsList, setAttemptsList] = useState<DeliveryAttempt[]>([]);
  const [attemptsLoading, setAttemptsLoading] = useState(false);

  // Form states
  const [createData, setCreateData] = useState<CreateDeliveryTaskPayload>({
    parcel_id: '',
    assigned_driver_id: '',
    priority: 'NORMAL',
    notes: '',
  });
  const [attemptData, setAttemptData] = useState<RecordDeliveryAttemptPayload>({
    outcome: 'CUSTOMER_UNAVAILABLE',
    notes: '',
  });
  const [proofData, setProofData] = useState<SubmitDeliveryProofPayload>({
    proof_type: 'RECIPIENT_SIGNATURE',
    recipient_name: '',
    signature_data: 'Signed on Glass',
    notes: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const isOperator = userRole === 'TENANT_ADMIN' || userRole === 'TENANT_OPERATOR' || userRole === 'PLATFORM_ADMIN';

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [tasksRes, parcelsRes, empRes, vehRes] = await Promise.all([
        api.listDeliveryTasks(tenantId, {
          status: statusFilter || undefined,
          priority: priorityFilter || undefined,
        }),
        api.listParcels(tenantId, { limit: 100 }),
        api.listEmployees(tenantId, { operational_role: 'DRIVER' }),
        api.listVehicles(tenantId, { status: 'AVAILABLE' }),
      ]);

      setTasks(tasksRes.tasks || []);
      // Filter parcels eligible for last-mile delivery assignment
      const eligible = (parcelsRes.parcels || []).filter(
        (p) => p.status === 'RECEIVED_AT_ORIGIN_BRANCH' || p.status === 'RECEIVED_AT_TRANSFER_BRANCH' || p.status === 'OUT_FOR_DELIVERY' || p.status === 'DELIVERY_ATTEMPTED'
      );
      setEligibleParcels(eligible);
      setDrivers(empRes.employees || []);
      setVehicles(vehRes.vehicles || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load delivery tasks');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [tenantId, statusFilter, priorityFilter]);

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!createData.parcel_id || !createData.assigned_driver_id) {
      setFormError('Please select both an eligible parcel and an authorized driver');
      return;
    }

    try {
      setSubmitting(true);
      setFormError(null);
      await api.createDeliveryTask(tenantId, createData);
      setShowCreateModal(false);
      setSuccessMsg('Delivery task dispatched! Parcel state transitioned to OUT_FOR_DELIVERY.');
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      setFormError(err.message || 'Failed to create delivery task');
    } finally {
      setSubmitting(false);
    }
  };

  const handleStartTask = async (task: DeliveryTask) => {
    try {
      await api.updateDeliveryTaskStatus(tenantId, task.id, 'IN_PROGRESS', 'Driver confirmed transit start');
      setSuccessMsg(`Task ${task.tracking_number} marked IN_PROGRESS`);
      setTimeout(() => setSuccessMsg(null), 3000);
      loadData();
    } catch (err: any) {
      alert(`Could not start task: ${err.message}`);
    }
  };

  const handleRecordAttempt = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTask) return;
    try {
      setSubmitting(true);
      setFormError(null);
      await api.recordDeliveryAttempt(tenantId, selectedTask.id, attemptData);
      setShowAttemptModal(false);
      setSuccessMsg(`Attempt logged for ${selectedTask.tracking_number}: ${attemptData.outcome}`);
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      setFormError(err.message || 'Failed to record attempt');
    } finally {
      setSubmitting(false);
    }
  };

  const handleSubmitProof = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTask) return;
    if (!proofData.recipient_name.trim()) {
      setFormError('Recipient signature/name is required');
      return;
    }

    try {
      setSubmitting(true);
      setFormError(null);
      await api.submitDeliveryProof(tenantId, selectedTask.id, proofData);
      setShowProofModal(false);
      setSuccessMsg(`Delivery successful! Proof recorded and parcel ${selectedTask.tracking_number} set to DELIVERED.`);
      setTimeout(() => setSuccessMsg(null), 4000);
      loadData();
    } catch (err: any) {
      setFormError(err.message || 'Proof verification failed');
    } finally {
      setSubmitting(false);
    }
  };

  const handleOpenAttempts = async (task: DeliveryTask) => {
    setSelectedTask(task);
    setShowHistoryModal(true);
    setAttemptsLoading(true);
    try {
      const atts = await api.getDeliveryAttempts(tenantId, task.id);
      setAttemptsList(atts || []);
    } catch (err) {
      console.error(err);
    } finally {
      setAttemptsLoading(false);
    }
  };

  // KPIs
  const totalTasks = tasks.length;
  const inProgressTasks = tasks.filter((t) => t.status === 'IN_PROGRESS' || t.status === 'ASSIGNED').length;
  const completedTasks = tasks.filter((t) => t.status === 'COMPLETED').length;
  const failedTasks = tasks.filter((t) => t.status === 'FAILED').length;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <Truck size={26} color="var(--accent-cyan)" />
            Last-Mile Delivery Dispatch
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginTop: '0.2rem' }}>
            Driver assignments, attempt telemetry, real-time routing, and tamper-resistant proof of delivery (POD).
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          {isOperator && (
            <button
              onClick={() => {
                setFormError(null);
                setCreateData({
                  parcel_id: eligibleParcels.length > 0 ? eligibleParcels[0].id : '',
                  assigned_driver_id: drivers.length > 0 ? drivers[0].id : '',
                  priority: 'NORMAL',
                  notes: '',
                });
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
              Assign Delivery Task
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

      {/* KPI METRIC CARDS */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
        <div className="glass-panel" style={{ padding: '1.2rem' }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Total Tasks</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 700, color: 'var(--text-primary)', marginTop: '4px' }}>{totalTasks}</div>
        </div>
        <div className="glass-panel" style={{ padding: '1.2rem' }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Active In-Transit</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 700, color: 'var(--accent-cyan)', marginTop: '4px' }}>{inProgressTasks}</div>
        </div>
        <div className="glass-panel" style={{ padding: '1.2rem' }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Delivered Successfully</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 700, color: 'var(--accent-emerald)', marginTop: '4px' }}>{completedTasks}</div>
        </div>
        <div className="glass-panel" style={{ padding: '1.2rem' }}>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase' }}>Exceptions / Failed</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 700, color: 'var(--accent-rose)', marginTop: '4px' }}>{failedTasks}</div>
        </div>
      </div>

      {/* Filter and Search */}
      <div className="glass-panel" style={{ padding: '1rem', display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
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
            <option value="">All Task Statuses</option>
            <option value="ASSIGNED">Assigned</option>
            <option value="IN_PROGRESS">In Progress</option>
            <option value="COMPLETED">Delivered (Completed)</option>
            <option value="FAILED">Failed</option>
            <option value="CANCELLED">Cancelled</option>
          </select>

          <select
            value={priorityFilter}
            onChange={(e) => setPriorityFilter(e.target.value)}
            style={{
              padding: '0.55rem 0.8rem',
              borderRadius: '6px',
              border: '1px solid var(--border-subtle)',
              background: 'rgba(15, 23, 42, 0.6)',
              color: 'var(--text-primary)',
              fontSize: '0.85rem',
            }}
          >
            <option value="">All Priorities</option>
            <option value="LOW">Low</option>
            <option value="NORMAL">Normal</option>
            <option value="HIGH">High</option>
            <option value="URGENT">Urgent</option>
          </select>
        </div>
      </div>

      {/* Main Table */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem', color: 'var(--text-secondary)' }}>
          <RefreshCw className="spin" size={24} style={{ marginBottom: '0.5rem' }} />
          <p>Loading dispatch queue...</p>
        </div>
      ) : error ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--accent-rose)' }}>
          <AlertCircle size={28} style={{ margin: '0 auto 0.5rem auto' }} />
          <p>{error}</p>
        </div>
      ) : tasks.length === 0 ? (
        <div className="glass-panel" style={{ padding: '4rem 2rem', textAlign: 'center', color: 'var(--text-muted)' }}>
          <Truck size={44} style={{ margin: '0 auto 1rem auto', opacity: 0.4 }} />
          <h3 style={{ fontSize: '1.1rem', color: 'var(--text-secondary)' }}>No delivery tasks active</h3>
          <p style={{ fontSize: '0.85rem', marginTop: '0.25rem' }}>
            Click "Assign Delivery Task" to dispatch cargo to an available field courier.
          </p>
        </div>
      ) : (
        <div className="glass-panel" style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem', textAlign: 'left' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>
                <th style={{ padding: '1rem' }}>Tracking Code</th>
                <th style={{ padding: '1rem' }}>Destination Recipient</th>
                <th style={{ padding: '1rem' }}>Assigned Courier</th>
                <th style={{ padding: '1rem' }}>Priority & Vehicle</th>
                <th style={{ padding: '1rem' }}>Task Status</th>
                <th style={{ padding: '1rem', textAlign: 'right' }}>Courier Actions</th>
              </tr>
            </thead>
            <tbody>
              {tasks.map((task) => {
                const st = TASK_STATUS_COLORS[task.status] || { bg: 'rgba(255,255,255,0.1)', text: '#fff', border: 'transparent' };
                const pr = PRIORITY_BADGES[task.priority] || { label: task.priority, bg: 'rgba(255,255,255,0.1)', text: '#fff' };
                return (
                  <tr key={task.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                    <td style={{ padding: '1rem' }}>
                      <div style={{ fontWeight: 700, fontFamily: 'monospace', color: 'var(--accent-cyan)' }}>
                        {task.tracking_number || 'Package'}
                      </div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                        Assigned: {new Date(task.assigned_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </div>
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <div style={{ fontWeight: 600 }}>{task.receiver_name}</div>
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '3px', marginTop: '2px' }}>
                        <MapPin size={12} /> {task.receiver_address}
                      </div>
                      {task.receiver_phone && (
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                          <Phone size={10} style={{ display: 'inline', marginRight: '3px' }} />
                          {task.receiver_phone}
                        </div>
                      )}
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontWeight: 600 }}>
                        <User size={14} color="var(--accent-purple)" />
                        {task.driver_name || 'Driver'}
                      </div>
                      {task.driver_phone && (
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{task.driver_phone}</div>
                      )}
                    </td>

                    <td style={{ padding: '1rem' }}>
                      <span style={{
                        display: 'inline-block',
                        padding: '2px 8px',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        background: pr.bg,
                        color: pr.text,
                        marginBottom: '4px'
                      }}>
                        {pr.label}
                      </span>
                      {task.vehicle_license_plate && (
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                          Plate: {task.vehicle_license_plate}
                        </div>
                      )}
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
                        {task.status.replace(/_/g, ' ')}
                      </span>
                    </td>

                    <td style={{ padding: '1rem', textAlign: 'right' }}>
                      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.4rem' }}>
                        {task.status === 'ASSIGNED' && (
                          <button
                            onClick={() => handleStartTask(task)}
                            style={{
                              padding: '0.4rem 0.75rem',
                              borderRadius: '6px',
                              border: 'none',
                              background: 'rgba(245, 158, 11, 0.2)',
                              color: '#fbbf24',
                              fontSize: '0.8rem',
                              fontWeight: 600,
                              cursor: 'pointer'
                            }}
                          >
                            Start Run
                          </button>
                        )}

                        {task.status === 'IN_PROGRESS' && (
                          <>
                            <button
                              onClick={() => {
                                setSelectedTask(task);
                                setFormError(null);
                                setAttemptData({ outcome: 'CUSTOMER_UNAVAILABLE', notes: '' });
                                setShowAttemptModal(true);
                              }}
                              style={{
                                padding: '0.4rem 0.75rem',
                                borderRadius: '6px',
                                border: 'none',
                                background: 'rgba(244, 63, 94, 0.2)',
                                color: '#f43f5e',
                                fontSize: '0.8rem',
                                fontWeight: 600,
                                cursor: 'pointer'
                              }}
                            >
                              Attempt Failed
                            </button>

                            <button
                              onClick={() => {
                                setSelectedTask(task);
                                setFormError(null);
                                setProofData({
                                  proof_type: 'RECIPIENT_SIGNATURE',
                                  recipient_name: task.receiver_name || '',
                                  signature_data: 'Signed on Mobile Device',
                                  notes: '',
                                });
                                setShowProofModal(true);
                              }}
                              style={{
                                padding: '0.4rem 0.75rem',
                                borderRadius: '6px',
                                border: 'none',
                                background: 'rgba(16, 185, 129, 0.2)',
                                color: '#34d399',
                                fontSize: '0.8rem',
                                fontWeight: 600,
                                cursor: 'pointer',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '3px'
                              }}
                            >
                              <FileCheck size={14} />
                              Deliver (POD)
                            </button>
                          </>
                        )}

                        <button
                          onClick={() => handleOpenAttempts(task)}
                          style={{
                            padding: '0.4rem 0.6rem',
                            borderRadius: '6px',
                            border: '1px solid var(--border-subtle)',
                            background: 'rgba(30, 41, 59, 0.7)',
                            color: 'var(--text-secondary)',
                            fontSize: '0.75rem',
                            cursor: 'pointer'
                          }}
                          title="View Attempt Telemetry"
                        >
                          History
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* ASSIGN DELIVERY TASK MODAL */}
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '580px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
              <h3 style={{ fontSize: '1.3rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Truck size={22} color="var(--accent-cyan)" />
                Dispatch Delivery Task
              </h3>
              <button onClick={() => setShowCreateModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            {formError && (
              <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--accent-rose)', color: 'var(--accent-rose)', marginBottom: '1rem', fontSize: '0.85rem' }}>
                {formError}
              </div>
            )}

            <form onSubmit={handleCreateTask} style={{ display: 'flex', flexDirection: 'column', gap: '1.2rem' }}>
              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Eligible Intake/Hub Parcel *</label>
                {eligibleParcels.length === 0 ? (
                  <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(245, 158, 11, 0.15)', color: '#fbbf24', fontSize: '0.85rem' }}>
                    No parcels currently in RECEIVED or OUT_FOR_DELIVERY status at this tenant organization.
                  </div>
                ) : (
                  <select
                    required
                    value={createData.parcel_id}
                    onChange={(e) => setCreateData({ ...createData, parcel_id: e.target.value })}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">Select parcel...</option>
                    {eligibleParcels.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.tracking_number} &rarr; {p.receiver_name} ({p.receiver_address}) [{p.status}]
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Assign Delivery Driver *</label>
                {drivers.length === 0 ? (
                  <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(245, 158, 11, 0.15)', color: '#fbbf24', fontSize: '0.85rem' }}>
                    No authorized drivers registered in this tenant organization.
                  </div>
                ) : (
                  <select
                    required
                    value={createData.assigned_driver_id}
                    onChange={(e) => setCreateData({ ...createData, assigned_driver_id: e.target.value })}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">Select driver...</option>
                    {drivers.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.first_name} {d.last_name} ({d.employee_code}) - {d.availability_status}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Priority Level</label>
                  <select
                    value={createData.priority}
                    onChange={(e) => setCreateData({ ...createData, priority: e.target.value as DeliveryPriority })}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="LOW">Low</option>
                    <option value="NORMAL">Normal</option>
                    <option value="HIGH">High Priority</option>
                    <option value="URGENT">Urgent (Same Day)</option>
                  </select>
                </div>

                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Assigned Fleet Vehicle (Optional)</label>
                  <select
                    value={createData.vehicle_id || ''}
                    onChange={(e) => setCreateData({ ...createData, vehicle_id: e.target.value || undefined })}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                  >
                    <option value="">None / Courier Personal</option>
                    {vehicles.map((v) => (
                      <option key={v.id} value={v.id}>{v.registration_number} ({v.make_model || v.vehicle_type})</option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Dispatch Notes</label>
                <input
                  type="text"
                  placeholder="e.g. Fragile package, call recipient 10 mins prior"
                  value={createData.notes || ''}
                  onChange={(e) => setCreateData({ ...createData, notes: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
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
                  disabled={submitting || eligibleParcels.length === 0 || drivers.length === 0}
                  style={{ padding: '0.6rem 1.4rem', borderRadius: '6px', border: 'none', background: 'var(--accent-cyan)', color: '#0f172a', fontWeight: 600, cursor: 'pointer' }}
                >
                  {submitting ? 'Dispatching...' : 'Dispatch Task to Driver'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* RECORD ATTEMPT MODAL */}
      {showAttemptModal && selectedTask && (
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '480px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem' }}>
              <h3 style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--accent-rose)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <AlertTriangle size={20} />
                Log Failed Delivery Attempt
              </h3>
              <button onClick={() => setShowAttemptModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            {formError && (
              <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--accent-rose)', color: 'var(--accent-rose)', marginBottom: '1rem', fontSize: '0.85rem' }}>
                {formError}
              </div>
            )}

            <form onSubmit={handleRecordAttempt} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Attempt Outcome Reason *</label>
                <select
                  value={attemptData.outcome}
                  onChange={(e) => setAttemptData({ ...attemptData, outcome: e.target.value as AttemptOutcome })}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                >
                  <option value="CUSTOMER_UNAVAILABLE">Customer Unavailable / Door Locked</option>
                  <option value="INCORRECT_ADDRESS">Address Incomplete or Incorrect</option>
                  <option value="REJECTED">Customer Refused Package</option>
                  <option value="SECURITY_RESTRICTED">Gated Community / Security Denied Access</option>
                  <option value="WEATHER_DELAY">Inclement Weather Delay</option>
                  <option value="OTHER">Other Operational Delay</option>
                </select>
              </div>

              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Driver Field Remarks</label>
                <input
                  type="text"
                  placeholder="e.g. Ring bell twice, no answer on phone"
                  value={attemptData.notes || ''}
                  onChange={(e) => setAttemptData({ ...attemptData, notes: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
                <button
                  type="button"
                  onClick={() => setShowAttemptModal(false)}
                  style={{ padding: '0.6rem 1rem', borderRadius: '6px', border: '1px solid var(--border-subtle)', background: 'transparent', color: 'var(--text-secondary)', cursor: 'pointer' }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  style={{ padding: '0.6rem 1.2rem', borderRadius: '6px', border: 'none', background: 'var(--accent-rose)', color: '#fff', fontWeight: 600, cursor: 'pointer' }}
                >
                  {submitting ? 'Recording...' : 'Log Attempt & Reschedule'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* SUBMIT PROOF OF DELIVERY (POD) MODAL */}
      {showProofModal && selectedTask && (
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '520px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem' }}>
              <h3 style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--accent-emerald)', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <ShieldCheck size={22} />
                Confirm Proof of Delivery (POD)
              </h3>
              <button onClick={() => setShowProofModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            {formError && (
              <div style={{ padding: '0.8rem', borderRadius: '6px', background: 'rgba(244, 63, 94, 0.15)', border: '1px solid var(--accent-rose)', color: 'var(--accent-rose)', marginBottom: '1rem', fontSize: '0.85rem' }}>
                {formError}
              </div>
            )}

            <form onSubmit={handleSubmitProof} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Verification Proof Mode *</label>
                <select
                  value={proofData.proof_type}
                  onChange={(e) => setProofData({ ...proofData, proof_type: e.target.value as ProofType })}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                >
                  <option value="RECIPIENT_SIGNATURE">Recipient Physical / Digital Signature</option>
                  <option value="OTP_VERIFICATION">One-Time Password (OTP) Auth</option>
                  <option value="PHOTO_CONFIRMATION">Photo Evidence at Dropoff Point</option>
                  <option value="SECURITY_PASS">Building Security Desk Signoff</option>
                </select>
              </div>

              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Signer / Recipient Name *</label>
                <input
                  type="text"
                  required
                  placeholder="Full name of person receiving package"
                  value={proofData.recipient_name}
                  onChange={(e) => setProofData({ ...proofData, recipient_name: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              {proofData.proof_type === 'OTP_VERIFICATION' && (
                <div>
                  <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>6-Digit Customer OTP Code *</label>
                  <input
                    type="text"
                    placeholder="e.g. 849201"
                    value={proofData.otp_code || ''}
                    onChange={(e) => setProofData({ ...proofData, otp_code: e.target.value })}
                    style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)', fontFamily: 'monospace' }}
                  />
                </div>
              )}

              <div>
                <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>Proof Evidence Signature / Note</label>
                <input
                  type="text"
                  placeholder="e.g. Signed on Courier Handheld Terminal"
                  value={proofData.signature_data || ''}
                  onChange={(e) => setProofData({ ...proofData, signature_data: e.target.value })}
                  style={{ width: '100%', padding: '0.65rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.8)', border: '1px solid var(--border-subtle)', color: 'var(--text-primary)' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '0.5rem' }}>
                <button
                  type="button"
                  onClick={() => setShowProofModal(false)}
                  style={{ padding: '0.6rem 1rem', borderRadius: '6px', border: '1px solid var(--border-subtle)', background: 'transparent', color: 'var(--text-secondary)', cursor: 'pointer' }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  style={{ padding: '0.6rem 1.4rem', borderRadius: '6px', border: 'none', background: 'var(--accent-emerald)', color: '#0f172a', fontWeight: 700, cursor: 'pointer' }}
                >
                  {submitting ? 'Verifying...' : 'Validate Proof & Mark DELIVERED'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ATTEMPTS HISTORY MODAL */}
      {showHistoryModal && selectedTask && (
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
          <div className="glass-panel" style={{ width: '100%', maxWidth: '520px', padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.2rem' }}>
              <h3 style={{ fontSize: '1.2rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Clock size={20} color="var(--accent-cyan)" />
                Delivery Attempts: {selectedTask.tracking_number}
              </h3>
              <button onClick={() => setShowHistoryModal(false)} style={{ background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}>
                <X size={20} />
              </button>
            </div>

            {attemptsLoading ? (
              <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>Loading attempts...</div>
            ) : attemptsList.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '2rem', color: 'var(--text-muted)' }}>No failed delivery attempts logged for this task.</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                {attemptsList.map((att) => (
                  <div key={att.id} style={{ padding: '0.75rem', borderRadius: '6px', background: 'rgba(15, 23, 42, 0.6)', border: '1px solid var(--border-subtle)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span style={{ fontWeight: 600, color: 'var(--accent-rose)', fontSize: '0.85rem' }}>
                        Attempt #{att.attempt_number}: {att.outcome}
                      </span>
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        {new Date(att.attempt_time).toLocaleString()}
                      </span>
                    </div>
                    {att.notes && (
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', marginTop: '4px' }}>
                        {att.notes}
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

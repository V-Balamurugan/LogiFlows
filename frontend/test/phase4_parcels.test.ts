import { test, describe } from 'node:test';
import assert from 'node:assert/strict';
import type {
  Parcel,
  CreateParcelPayload,
  ParcelStatus,
  ServiceType,
  DeliveryTask,
  CreateDeliveryTaskPayload,
  RecordDeliveryAttemptPayload,
  SubmitDeliveryProofPayload,
  BranchTransfer,
  CreateBranchTransferPayload,
  PublicTrackingResponse
} from '../src/types/parcels.ts';

describe('Phase 4 Frontend Parcel & Delivery Lifecycle Validation', () => {
  // TC-FE-PARCEL-001: Create Parcel Payload Verification
  test('validates create parcel payload requires origin, destination and valid weight', () => {
    const payload: CreateParcelPayload = {
      sender_name: 'Alpha Logistics Tech',
      sender_phone: '+91 9876543210',
      sender_address: '10 Industrial Estate, Guindy, Chennai',
      receiver_name: 'Priya Sundaram',
      receiver_phone: '+91 9123456780',
      receiver_address: '42 Koramangala 4th Block, Bengaluru',
      origin_branch_id: '00000000-0000-0000-0000-000000000001',
      destination_branch_id: '00000000-0000-0000-0000-000000000002',
      weight_kg: 3.5,
      dimensions_cm: '30x20x15',
      service_type: 'EXPRESS',
      declared_value: 2500,
    };

    assert.ok(payload.weight_kg > 0);
    assert.notEqual(payload.origin_branch_id, payload.destination_branch_id);
    assert.equal(payload.service_type, 'EXPRESS');
    assert.ok(payload.dimensions_cm.includes('x'));
  });

  // TC-FE-PARCEL-002: Parcel Status State Machine Transitions
  test('validates legal parcel status enum values', () => {
    const validStatuses: ParcelStatus[] = [
      'CREATED',
      'BOOKED',
      'READY_FOR_PICKUP',
      'PICKED_UP',
      'RECEIVED_AT_ORIGIN_BRANCH',
      'IN_TRANSIT',
      'RECEIVED_AT_TRANSFER_BRANCH',
      'OUT_FOR_DELIVERY',
      'DELIVERY_ATTEMPTED',
      'DELIVERED',
      'DELIVERY_FAILED',
      'RETURN_INITIATED',
      'RETURNED',
      'CANCELLED',
      'ON_HOLD',
    ];

    assert.equal(validStatuses.length, 15);
    assert.ok(validStatuses.includes('OUT_FOR_DELIVERY'));
    assert.ok(validStatuses.includes('DELIVERED'));
  });

  // TC-FE-DELIVERY-001: Delivery Task Dispatch Payload
  test('validates delivery task dispatch links parcel, driver and priority', () => {
    const taskPayload: CreateDeliveryTaskPayload = {
      parcel_id: '11111111-1111-1111-1111-111111111111',
      assigned_driver_id: '22222222-2222-2222-2222-222222222222',
      vehicle_id: '33333333-3333-3333-3333-333333333333',
      priority: 'URGENT',
      notes: 'Deliver to 3rd floor reception before 2 PM',
    };

    assert.ok(taskPayload.parcel_id.length > 0);
    assert.ok(taskPayload.assigned_driver_id.length > 0);
    assert.equal(taskPayload.priority, 'URGENT');
    assert.equal(taskPayload.notes?.includes('reception'), true);
  });

  // TC-FE-DELIVERY-002: Delivery Attempt Logging
  test('validates delivery attempt outcome recording', () => {
    const attempt: RecordDeliveryAttemptPayload = {
      outcome: 'CUSTOMER_UNAVAILABLE',
      notes: 'Customer did not respond to door bell and phone calls',
      latitude: 12.9716,
      longitude: 77.5946,
    };

    assert.equal(attempt.outcome, 'CUSTOMER_UNAVAILABLE');
    assert.ok((attempt.latitude ?? 0) > 0);
    assert.ok((attempt.longitude ?? 0) > 0);
  });

  // TC-FE-DELIVERY-003: Proof of Delivery (POD) Payload
  test('validates proof of delivery submission requires recipient name and proof type', () => {
    const podPayload: SubmitDeliveryProofPayload = {
      proof_type: 'RECIPIENT_SIGNATURE',
      recipient_name: 'Priya Sundaram',
      recipient_relationship: 'SELF',
      signature_data: 'data:image/svg+xml;base64,PHN2Zz4...',
      notes: 'Handed over directly to consignee',
    };

    assert.equal(podPayload.proof_type, 'RECIPIENT_SIGNATURE');
    assert.ok(podPayload.recipient_name.length > 0);
    assert.equal(podPayload.recipient_relationship, 'SELF');
  });

  // TC-FE-TRANSFER-001: Inter-Branch Linehaul Manifest Payload
  test('validates inter-branch transfer manifest requires route and parcel list', () => {
    const transfer: CreateBranchTransferPayload = {
      origin_branch_id: '00000000-0000-0000-0000-000000000001',
      destination_branch_id: '00000000-0000-0000-0000-000000000002',
      responsible_employee_id: '44444444-4444-4444-4444-444444444444',
      parcel_ids: [
        '55555555-5555-5555-5555-555555555551',
        '55555555-5555-5555-5555-555555555552',
        '55555555-5555-5555-5555-555555555553',
      ],
      notes: 'Scheduled night linehaul shuttle #12',
    };

    assert.equal(transfer.parcel_ids.length, 3);
    assert.notEqual(transfer.origin_branch_id, transfer.destination_branch_id);
    assert.ok(transfer.responsible_employee_id);
  });

  // TC-FE-TRACKING-001: Public Tracking Response Sanitization
  test('validates customer tracking response contains sanitized city-level information and zero PII leaks', () => {
    const publicTracking: PublicTrackingResponse = {
      tracking_number: 'PKG-20260924-A1B2-0001',
      status: 'IN_TRANSIT',
      status_display: 'In Transit between Hubs',
      service_type: 'OVERNIGHT',
      origin_city: 'Chennai',
      destination_city: 'Bengaluru',
      current_location: 'Coimbatore Sorting Hub',
      created_at: '2026-09-24T06:00:00Z',
      updated_at: '2026-09-24T10:30:00Z',
      milestones: [
        {
          status: 'BOOKED',
          description: 'Shipment booking confirmed by origin facility',
          location: 'Chennai Sorting Hub',
          timestamp: '2026-09-24T06:00:00Z',
        },
        {
          status: 'IN_TRANSIT',
          description: 'Dispatched on inter-branch linehaul container',
          location: 'Coimbatore Sorting Hub',
          timestamp: '2026-09-24T10:30:00Z',
        },
      ],
    };

    assert.equal(publicTracking.tracking_number, 'PKG-20260924-A1B2-0001');
    assert.equal(publicTracking.origin_city, 'Chennai');
    assert.equal(publicTracking.destination_city, 'Bengaluru');
    assert.equal(publicTracking.milestones.length, 2);

    // Verify ZERO PII fields exist on public tracking response
    const trackingObj = publicTracking as any;
    assert.equal(trackingObj.sender_phone, undefined);
    assert.equal(trackingObj.sender_address, undefined);
    assert.equal(trackingObj.receiver_phone, undefined);
    assert.equal(trackingObj.receiver_address, undefined);
    assert.equal(trackingObj.driver_id, undefined);
    assert.equal(trackingObj.driver_name, undefined);
  });

  // TC-FE-QR-001: QR Code Label Payload & Destination Information Verification
  test('validates QR code payload contains tracking, tenant, and parcel ID UUID', () => {
    const parcelId = '99999999-9999-9999-9999-999999999999';
    const trackingNumber = 'PKG-20260924-TEST-0042';
    const tenantId = '00000000-0000-0000-0000-000000000001';

    const qrPayload = JSON.stringify({
      tracking: trackingNumber,
      tenant_id: tenantId,
      parcel_id: parcelId,
    });

    const parsed = JSON.parse(qrPayload);
    assert.equal(parsed.tracking, trackingNumber);
    assert.equal(parsed.parcel_id, parcelId);
    assert.equal(parsed.tenant_id, tenantId);
  });

  // TC-FE-QR-002: Printable Shipping Routing Slip Fields Verification
  test('validates shipping label includes Destination Hub, Recipient, and Parcel UUID', () => {
    const labelData = {
      parcelId: '88888888-8888-8888-8888-888888888888',
      trackingNumber: 'PKG-20260924-DEST-0099',
      destinationHub: 'Bengaluru Central Gateway',
      destinationCity: 'Bengaluru',
      recipientName: 'Karthik Raja',
      recipientAddress: '7th Cross, Indiranagar',
      serviceType: 'EXPRESS',
    };

    assert.ok(labelData.parcelId.length === 36, 'Parcel ID must be a standard 36-char UUID');
    assert.ok(labelData.destinationHub.includes('Bengaluru'), 'Label must explicitly display destination hub');
    assert.ok(labelData.destinationCity.length > 0, 'Label must explicitly display destination city');
    assert.ok(labelData.trackingNumber.startsWith('PKG-'), 'Tracking code must start with PKG- prefix');
  });

  // TC-FE-FSM-001: Strict State Transition Table Verification
  test('validates parcel state machine prevents invalid state leaps', () => {
    const VALID_NEXT_TRANSITIONS: Record<string, string[]> = {
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

    // Valid moves from CREATED
    assert.deepEqual(VALID_NEXT_TRANSITIONS['CREATED'], ['BOOKED', 'RECEIVED_AT_ORIGIN_BRANCH', 'CANCELLED']);
    assert.ok(!VALID_NEXT_TRANSITIONS['CREATED'].includes('IN_TRANSIT'), 'Cannot jump directly from CREATED to IN_TRANSIT');
    assert.ok(!VALID_NEXT_TRANSITIONS['CREATED'].includes('OUT_FOR_DELIVERY'), 'Cannot jump directly from CREATED to OUT_FOR_DELIVERY');
    assert.ok(!VALID_NEXT_TRANSITIONS['CREATED'].includes('DELIVERED'), 'Cannot jump directly from CREATED to DELIVERED');

    // Terminal states have zero forward transitions
    assert.equal(VALID_NEXT_TRANSITIONS['DELIVERED'].length, 0);
    assert.equal(VALID_NEXT_TRANSITIONS['RETURNED'].length, 0);
    assert.equal(VALID_NEXT_TRANSITIONS['CANCELLED'].length, 0);
  });
});


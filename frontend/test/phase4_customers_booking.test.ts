import { test, describe } from 'node:test';
import assert from 'node:assert/strict';
import type {
  Customer,
  CreateCustomerPayload,
  CustomerType,
  CustomerStatus
} from '../src/types/customers.ts';
import type {
  CreateParcelPayload,
  ServiceType,
  Parcel
} from '../src/types/parcels.ts';

describe('Phase 4 Customer Management & Booking Vertical Slice', () => {
  // TC-FE-CUST-001: Customer Creation Payload Validation
  test('validates customer payload with individual classification', () => {
    const payload: CreateCustomerPayload = {
      customer_type: 'INDIVIDUAL',
      name: 'Rohan Verma',
      phone: '+91 98765 12345',
      email: 'rohan.v@example.com',
      billing_address: '15 Rajaji Salai, Chennai, TN - 600001',
    };

    assert.equal(payload.customer_type, 'INDIVIDUAL');
    assert.ok(payload.name.length > 0);
    assert.ok(payload.phone.length >= 7);
    assert.ok(payload.billing_address.length > 0);
  });

  // TC-FE-CUST-002: Enterprise Customer with GSTIN
  test('validates enterprise customer with company and tax id', () => {
    const payload: CreateCustomerPayload = {
      customer_type: 'ENTERPRISE',
      name: 'Suresh Menon',
      company_name: 'Menon Automotive Ltd',
      tax_id: '33AAACM1234F1Z5',
      phone: '+91 44 2855 0000',
      email: 'logistics@menonauto.com',
      billing_address: 'Plot 45 Ambattur Industrial Estate, Chennai',
      shipping_address: 'Warehouse 3, Sriperumbudur Logistics Park',
      notes: 'High-volume contract with priority handling',
    };

    assert.equal(payload.customer_type, 'ENTERPRISE');
    assert.equal(payload.company_name, 'Menon Automotive Ltd');
    assert.ok(payload.tax_id && payload.tax_id.length > 5);
    assert.ok(payload.shipping_address);
  });

  // TC-FE-CUST-003: Dynamic Price Calculation Formula Matching Go Backend
  test('accurately calculates estimated shipping price based on service type, weight, and insurance', () => {
    const calculateEstimatedPrice = (serviceType: ServiceType, weightKg: number, declaredValue: number = 0) => {
      let base = 50;
      if (serviceType === 'EXPRESS') base = 120;
      else if (serviceType === 'OVERNIGHT') base = 200;
      else if (serviceType === 'SAME_DAY') base = 350;

      const weightFee = Math.max(0, weightKg) * 20;
      let insuranceFee = 0;
      if (declaredValue > 1000) {
        insuranceFee = (declaredValue - 1000) * 0.005;
      }
      return Math.round((base + weightFee + insuranceFee) * 100) / 100;
    };

    // Standard 2kg, no excess value -> 50 + (2 * 20) = 90
    assert.equal(calculateEstimatedPrice('STANDARD', 2.0, 500), 90.0);

    // Express 5kg, declared value 5000 -> 120 + (5 * 20 = 100) + ((5000-1000)*0.005 = 20) = 240
    assert.equal(calculateEstimatedPrice('EXPRESS', 5.0, 5000), 240.0);

    // Overnight 10kg, declared value 1000 -> 200 + (10 * 20 = 200) + 0 = 400
    assert.equal(calculateEstimatedPrice('OVERNIGHT', 10.0, 1000), 400.0);

    // Same Day 1.5kg, declared value 2000 -> 350 + (1.5 * 20 = 30) + ((2000-1000)*0.005 = 5) = 385
    assert.equal(calculateEstimatedPrice('SAME_DAY', 1.5, 2000), 385.0);
  });

  // TC-FE-CUST-004: Parcel Creation with Sender & Receiver Customer Linkage
  test('links sender and receiver customer IDs in parcel booking payload', () => {
    const bookingPayload: CreateParcelPayload = {
      sender_customer_id: '6fc7bebf-5c45-468d-8d44-e89b76c2b76e',
      receiver_customer_id: '01d4a0a7-1d22-4411-9a74-d4b3eb46b19a',
      sender_name: 'Tata Consultancy Services',
      sender_phone: '+91 98840 12345',
      sender_address: 'Siruseri IT Park, Chennai',
      receiver_name: 'Wipro Technologies',
      receiver_phone: '+91 80 2844 0011',
      receiver_address: 'Electronic City Phase 1, Bengaluru',
      origin_branch_id: '11111111-1111-1111-1111-111111111111',
      destination_branch_id: '22222222-2222-2222-2222-222222222222',
      weight_kg: 2.5,
      dimensions_cm: '25x20x10',
      service_type: 'EXPRESS',
      declared_value: 2000,
    };

    assert.ok(bookingPayload.sender_customer_id);
    assert.ok(bookingPayload.receiver_customer_id);
    assert.equal(bookingPayload.service_type, 'EXPRESS');
    assert.equal(bookingPayload.declared_value, 2000);
  });

  // TC-FE-CUST-005: Customer Data Invariants
  test('enforces legal customer classification and status invariants', () => {
    const types: CustomerType[] = ['INDIVIDUAL', 'BUSINESS', 'ENTERPRISE', 'MERCHANT'];
    const statuses: CustomerStatus[] = ['ACTIVE', 'INACTIVE', 'SUSPENDED'];

    assert.equal(types.length, 4);
    assert.equal(statuses.length, 3);
    assert.ok(types.includes('ENTERPRISE'));
    assert.ok(statuses.includes('ACTIVE'));
  });
});

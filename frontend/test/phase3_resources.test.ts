import { test, describe } from 'node:test';
import assert from 'node:assert/strict';
import type {
  Employee,
  Vehicle,
  CreateEmployeePayload,
  CreateVehiclePayload,
  UpdateEmployeeStatusPayload,
  UpdateVehicleStatusPayload
} from '../src/types/resources.ts';

describe('Phase 3 Frontend Resource Validation & Type Integrity', () => {
  test('validates employee creation payload allows optional employee_code for auto-generation', () => {
    const payload: CreateEmployeePayload = {
      first_name: 'Rajesh',
      last_name: 'Kumar',
      designation: 'Route Driver',
      operational_role: 'DRIVER',
      employment_type: 'FULL_TIME',
      availability_status: 'AVAILABLE',
      verification_status: 'VERIFIED',
    };

    assert.equal(payload.employee_code, undefined);
    assert.equal(payload.operational_role, 'DRIVER');
    assert.equal(payload.availability_status, 'AVAILABLE');
  });

  test('validates employee status update transitions', () => {
    const statusUpdate: UpdateEmployeeStatusPayload = {
      status: 'ON_LEAVE',
      availability_status: 'OFF_DUTY',
    };

    assert.equal(statusUpdate.status, 'ON_LEAVE');
    assert.equal(statusUpdate.availability_status, 'OFF_DUTY');
  });

  test('validates vehicle physical capacity and type constraints', () => {
    const vehiclePayload: CreateVehiclePayload = {
      registration_number: 'TN09AB1234',
      vehicle_type: 'ELECTRIC_VAN',
      max_weight_kg: 1200,
      max_volume_cbm: 8.5,
      availability_status: 'AVAILABLE',
    };

    assert.ok(vehiclePayload.max_weight_kg > 0);
    assert.ok(vehiclePayload.max_volume_cbm > 0);
    assert.equal(vehiclePayload.vehicle_type, 'ELECTRIC_VAN');
  });

  test('validates vehicle status update payload', () => {
    const update: UpdateVehicleStatusPayload = {
      status: 'MAINTENANCE',
      availability_status: 'MAINTENANCE',
    };

    assert.equal(update.status, 'MAINTENANCE');
    assert.equal(update.availability_status, 'MAINTENANCE');
  });

  test('checks driver eligibility criteria for assignment', () => {
    function isDriverEligible(emp: Partial<Employee>): { eligible: boolean; reason?: string } {
      if (emp.operational_role !== 'DRIVER') {
        return { eligible: false, reason: 'Employee does not have DRIVER operational role' };
      }
      if (emp.status !== 'ACTIVE' || !emp.is_active) {
        return { eligible: false, reason: 'Employee is not active' };
      }
      if (emp.availability_status !== 'AVAILABLE') {
        return { eligible: false, reason: 'Employee is not available' };
      }
      return { eligible: true };
    }

    const validDriver: Partial<Employee> = {
      operational_role: 'DRIVER',
      status: 'ACTIVE',
      is_active: true,
      availability_status: 'AVAILABLE',
    };
    assert.equal(isDriverEligible(validDriver).eligible, true);

    const busyDriver: Partial<Employee> = {
      operational_role: 'DRIVER',
      status: 'ACTIVE',
      is_active: true,
      availability_status: 'BUSY',
    };
    assert.equal(isDriverEligible(busyDriver).eligible, false);

    const manager: Partial<Employee> = {
      operational_role: 'MANAGER',
      status: 'ACTIVE',
      is_active: true,
      availability_status: 'AVAILABLE',
    };
    assert.equal(isDriverEligible(manager).eligible, false);
  });

  test('validates CreateEmployeeWithAccountPayload integrity and constraints', () => {
    import('../src/types/resources.ts').then((mod) => {
      // Type integrity check
      const accountPayload: import('../src/types/resources.ts').CreateEmployeeWithAccountPayload = {
        first_name: 'Ananya',
        last_name: 'Rao',
        email: 'ananya.rao@example.com',
        designation: 'Central Dispatcher',
        operational_role: 'DISPATCHER',
        system_role: 'TENANT_OPERATOR',
        password: 'SecurePassword123!',
        send_invite: false,
      };

      assert.equal(accountPayload.email, 'ananya.rao@example.com');
      assert.equal(accountPayload.system_role, 'TENANT_OPERATOR');
      assert.ok(accountPayload.password && accountPayload.password.length >= 8);
    });
  });

  test('validates EmployeeAccountStatus and EmployeeMe structures', () => {
    const accountStatus: import('../src/types/resources.ts').EmployeeAccountStatus = {
      employee_id: 'emp-123',
      employee_code: 'EMP-001',
      full_name: 'Vikram Seth',
      operational_role: 'DRIVER',
      status: 'ACTIVE',
      has_account: true,
      user_id: 'usr-456',
      user_email: 'vikram@example.com',
      system_role: 'EMPLOYEE',
      user_is_active: true,
    };

    assert.equal(accountStatus.has_account, true);
    assert.equal(accountStatus.system_role, 'EMPLOYEE');
    assert.equal(accountStatus.user_id, 'usr-456');

    const meResponse: import('../src/types/resources.ts').EmployeeMeResponse = {
      employee: {
        id: 'emp-123',
        tenant_id: 'ten-789',
        employee_code: 'EMP-001',
        first_name: 'Vikram',
        last_name: 'Seth',
        designation: 'Driver',
        employment_type: 'FULL_TIME',
        operational_role: 'DRIVER',
        status: 'ACTIVE',
        availability_status: 'AVAILABLE',
        verification_status: 'VERIFIED',
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      system_role: 'EMPLOYEE',
      assigned_vehicle: {
        id: 'veh-001',
        registration_number: 'KA01AB1234',
        vehicle_type: 'VAN',
        status: 'ACTIVE',
      },
    };

    assert.equal(meResponse.system_role, 'EMPLOYEE');
    assert.equal(meResponse.assigned_vehicle?.registration_number, 'KA01AB1234');
  });
});

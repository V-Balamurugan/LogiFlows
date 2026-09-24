import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/models/resource_models.dart';

void main() {
  group('Resource Models Unit Tests (Phase 2)', () {
    test('BranchModel deserializes correctly from complete JSON payload', () {
      final json = {
        'id': 'b-1001',
        'tenant_id': 't-001',
        'name': 'North Chennai Distribution Hub',
        'branch_code': 'NCH-01',
        'address': 'Plot 45, Ambattur Industrial Estate',
        'city': 'Chennai',
        'state': 'Tamil Nadu',
        'postal_code': '600058',
        'country': 'IN',
        'latitude': 13.0827,
        'longitude': 80.2707,
        'coverage_radius_km': 35.5,
        'status': 'ACTIVE',
        'is_active': true,
        'created_at': '2026-09-22T10:00:00Z',
        'updated_at': '2026-09-22T10:00:00Z',
      };

      final branch = BranchModel.fromJson(json);

      expect(branch.id, 'b-1001');
      expect(branch.tenantId, 't-001');
      expect(branch.name, 'North Chennai Distribution Hub');
      expect(branch.branchCode, 'NCH-01');
      expect(branch.city, 'Chennai');
      expect(branch.latitude, 13.0827);
      expect(branch.longitude, 80.2707);
      expect(branch.coverageRadiusKm, 35.5);
      expect(branch.isActive, isTrue);
      expect(branch.status, 'ACTIVE');
    });

    test('BranchModel handles integer latitude and coverage radius smoothly', () {
      final json = {
        'id': 'b-1002',
        'tenant_id': 't-001',
        'name': 'Central Depot',
        'branch_code': 'CEN-01',
        'address': 'Main St',
        'city': 'Bengaluru',
        'latitude': 13,
        'longitude': 77,
        'coverage_radius_km': 25,
        'status': 'ACTIVE',
        'is_active': true,
      };

      final branch = BranchModel.fromJson(json);

      expect(branch.latitude, 13.0);
      expect(branch.longitude, 77.0);
      expect(branch.coverageRadiusKm, 25.0);
    });

    test('EmployeeModel deserializes operational driver with branch assignment', () {
      final json = {
        'id': 'emp-201',
        'tenant_id': 't-001',
        'user_id': 'usr-301',
        'branch_id': 'b-1001',
        'employee_code': 'EMP-DRV-001',
        'role': 'DRIVER',
        'full_name': 'Murugan V',
        'email': 'murugan.driver@logiflows.test',
        'phone': '+919876543210',
        'status': 'ACTIVE',
        'is_active': true,
        'branch_name': 'North Chennai Distribution Hub',
        'created_at': '2026-09-22T10:00:00Z',
        'updated_at': '2026-09-22T10:00:00Z',
      };

      final emp = EmployeeModel.fromJson(json);

      expect(emp.id, 'emp-201');
      expect(emp.role, 'DRIVER');
      expect(emp.fullName, 'Murugan V');
      expect(emp.branchName, 'North Chennai Distribution Hub');
      expect(emp.isActive, isTrue);
    });

    test('VehicleModel deserializes electric commercial cargo vehicle', () {
      final json = {
        'id': 'veh-501',
        'tenant_id': 't-001',
        'branch_id': 'b-1001',
        'registration_number': 'TN-05-EV-1234',
        'vehicle_type': 'ELECTRIC_VAN',
        'make_model': 'Tata Ace EV',
        'year': 2025,
        'max_weight_kg': 1000.0,
        'max_volume_cbm': 6.5,
        'is_electric': true,
        'status': 'AVAILABLE',
        'current_driver_id': null,
        'current_driver_name': null,
        'branch_name': 'North Chennai Distribution Hub',
      };

      final vehicle = VehicleModel.fromJson(json);

      expect(vehicle.id, 'veh-501');
      expect(vehicle.registrationNumber, 'TN-05-EV-1234');
      expect(vehicle.vehicleType, 'ELECTRIC_VAN');
      expect(vehicle.isElectric, isTrue);
      expect(vehicle.maxWeightKg, 1000.0);
      expect(vehicle.maxVolumeCbm, 6.5);
      expect(vehicle.status, 'AVAILABLE');
      expect(vehicle.currentDriverName, isNull);
    });

    test('VehicleModel deserializes assigned vehicle with driver details', () {
      final json = {
        'id': 'veh-502',
        'tenant_id': 't-001',
        'registration_number': 'TN-01-AB-9876',
        'vehicle_type': 'TRUCK',
        'make_model': 'Ashok Leyland Dost',
        'max_weight_kg': 2500,
        'max_volume_cbm': 12,
        'is_electric': false,
        'status': 'ASSIGNED',
        'current_driver_id': 'emp-201',
        'current_driver_name': 'Murugan V',
      };

      final vehicle = VehicleModel.fromJson(json);

      expect(vehicle.id, 'veh-502');
      expect(vehicle.status, 'ASSIGNED');
      expect(vehicle.currentDriverId, 'emp-201');
      expect(vehicle.currentDriverName, 'Murugan V');
      expect(vehicle.isElectric, isFalse);
      expect(vehicle.maxWeightKg, 2500.0);
    });

    test('EmployeeModel deserializes Phase 3 availability and verification statuses', () {
      final json = {
        'id': 'emp-301',
        'tenant_id': 't-001',
        'employee_code': 'EMP-0042',
        'first_name': 'Kavitha',
        'last_name': 'Raman',
        'operational_role': 'DRIVER',
        'availability_status': 'AVAILABLE',
        'verification_status': 'VERIFIED',
        'status': 'ACTIVE',
        'is_active': true,
      };

      final emp = EmployeeModel.fromJson(json);

      expect(emp.employeeCode, 'EMP-0042');
      expect(emp.fullName, 'Kavitha Raman');
      expect(emp.isDriver, isTrue);
      expect(emp.isAvailable, isTrue);
      expect(emp.availabilityStatus, 'AVAILABLE');
      expect(emp.verificationStatus, 'VERIFIED');
    });

    test('AssignmentModel deserializes active vehicle-driver link', () {
      final json = {
        'id': 'asgn-101',
        'tenant_id': 't-001',
        'vehicle_id': 'veh-501',
        'employee_id': 'emp-301',
        'registration_number': 'TN-05-EV-1234',
        'vehicle_type': 'ELECTRIC_VAN',
        'driver_name': 'Kavitha Raman',
        'driver_code': 'EMP-0042',
        'assigned_at': '2026-09-23T10:00:00Z',
        'status': 'ACTIVE',
        'notes': 'Morning dispatch run',
      };

      final asgn = AssignmentModel.fromJson(json);

      expect(asgn.id, 'asgn-101');
      expect(asgn.vehicleId, 'veh-501');
      expect(asgn.employeeId, 'emp-301');
      expect(asgn.registrationNumber, 'TN-05-EV-1234');
      expect(asgn.driverName, 'Kavitha Raman');
      expect(asgn.driverCode, 'EMP-0042');
      expect(asgn.isActive, isTrue);
    });

    test('EmployeeMeModel deserializes current user profile with assigned vehicle', () {
      final json = {
        'employee': {
          'id': 'emp-801',
          'tenant_id': 't-001',
          'employee_code': 'EMP-0088',
          'first_name': 'Deepak',
          'last_name': 'Sharma',
          'designation': 'Senior Fleet Driver',
          'operational_role': 'DRIVER',
          'status': 'ACTIVE',
          'availability_status': 'BUSY',
          'verification_status': 'VERIFIED',
          'is_active': true,
        },
        'system_role': 'EMPLOYEE',
        'assigned_vehicle': {
          'id': 'veh-901',
          'registration_number': 'MH-02-CB-4567',
          'vehicle_type': 'TRUCK',
          'make_model': 'Eicher Pro 2049',
          'status': 'ACTIVE',
        },
      };

      final me = EmployeeMeModel.fromJson(json);

      expect(me.employee.employeeCode, 'EMP-0088');
      expect(me.employee.fullName, 'Deepak Sharma');
      expect(me.systemRole, 'EMPLOYEE');
      expect(me.assignedVehicle, isNotNull);
      expect(me.assignedVehicle!.registrationNumber, 'MH-02-CB-4567');
      expect(me.assignedVehicle!.vehicleType, 'TRUCK');
    });

    test('EmployeeAccountStatusModel deserializes account status verification', () {
      final json = {
        'employee_id': 'emp-801',
        'employee_code': 'EMP-0088',
        'full_name': 'Deepak Sharma',
        'operational_role': 'DRIVER',
        'status': 'ACTIVE',
        'has_account': true,
        'user_id': 'usr-999',
        'user_email': 'deepak@logiflows.test',
        'user_is_active': true,
        'system_role': 'EMPLOYEE',
      };

      final status = EmployeeAccountStatusModel.fromJson(json);

      expect(status.employeeId, 'emp-801');
      expect(status.hasAccount, isTrue);
      expect(status.userEmail, 'deepak@logiflows.test');
      expect(status.systemRole, 'EMPLOYEE');
    });
  });
}

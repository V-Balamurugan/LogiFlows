import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/services/resource_api_client.dart';

void main() {
  group('ResourceApiClient Unit Tests (Phase 2)', () {
    late TokenStorage tokenStorage;

    setUp(() async {
      tokenStorage = InMemorySecureTokenStorage();
      await tokenStorage.clearTokens();
      await tokenStorage.saveTokens(accessToken: 'mock.valid.driver.jwt');
    });

    test('getBranches fetches and deserializes branch list with proximity params', () async {
      final mockClient = MockClient((request) async {
        expect(request.method, 'GET');
        expect(request.url.path, '/api/v1/tenants/tenant-100/branches');
        expect(request.url.queryParameters['near_lat'], '13.0827');
        expect(request.url.queryParameters['near_lng'], '80.2707');
        expect(request.url.queryParameters['radius_km'], '50.0');
        expect(request.headers['Authorization'], 'Bearer mock.valid.driver.jwt');

        final responseBody = jsonEncode({
          'success': true,
          'data': {
            'branches': [
              {
                'id': 'b-1',
                'tenant_id': 'tenant-100',
                'name': 'Ambattur Hub',
                'branch_code': 'AMB-01',
                'address': 'Estate Rd',
                'city': 'Chennai',
                'latitude': 13.0827,
                'longitude': 80.2707,
                'coverage_radius_km': 50.0,
                'status': 'ACTIVE',
                'is_active': true,
              }
            ],
            'total': 1,
          }
        });

        return http.Response(responseBody, 200, headers: {'content-type': 'application/json'});
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);
      final branches = await client.getBranches(
        'tenant-100',
        nearLat: 13.0827,
        nearLng: 80.2707,
        radiusKm: 50.0,
      );

      expect(branches.length, 1);
      expect(branches.first.name, 'Ambattur Hub');
      expect(branches.first.branchCode, 'AMB-01');
    });

    test('getVehicles fetches filtered vehicle fleet', () async {
      final mockClient = MockClient((request) async {
        expect(request.method, 'GET');
        expect(request.url.path, '/api/v1/tenants/tenant-100/vehicles');
        expect(request.url.queryParameters['vehicle_type'], 'VAN');
        expect(request.url.queryParameters['status'], 'AVAILABLE');

        final responseBody = jsonEncode({
          'success': true,
          'data': {
            'vehicles': [
              {
                'id': 'v-101',
                'tenant_id': 'tenant-100',
                'registration_number': 'TN-05-AB-1234',
                'vehicle_type': 'VAN',
                'is_electric': true,
                'max_weight_kg': 1200.0,
                'max_volume_cbm': 8.0,
                'status': 'AVAILABLE',
              }
            ],
            'total': 1,
          }
        });

        return http.Response(responseBody, 200, headers: {'content-type': 'application/json'});
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);
      final vehicles = await client.getVehicles(
        'tenant-100',
        vehicleType: 'VAN',
        status: 'AVAILABLE',
      );

      expect(vehicles.length, 1);
      expect(vehicles.first.registrationNumber, 'TN-05-AB-1234');
      expect(vehicles.first.isElectric, isTrue);
    });

    test('assignVehicle sends driver assignment and returns true on 201', () async {
      final mockClient = MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/v1/tenants/tenant-100/vehicles/v-101/assign');
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['driver_id'], 'drv-999');
        expect(body['notes'], 'Assigned for morning parcel route');

        return http.Response(
          jsonEncode({'success': true, 'data': {'assignment_id': 'asg-001'}}),
          201,
          headers: {'content-type': 'application/json'},
        );
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);
      final success = await client.assignVehicle(
        'tenant-100',
        'v-101',
        'drv-999',
        notes: 'Assigned for morning parcel route',
      );

      expect(success, isTrue);
    });

    test('assignVehicle throws conflict exception when 409 returned', () async {
      final mockClient = MockClient((request) async {
        return http.Response(
          jsonEncode({'success': false, 'error': {'code': 'RESOURCE_CONFLICT', 'message': 'vehicle already assigned'}}),
          409,
          headers: {'content-type': 'application/json'},
        );
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);

      expect(
        () => client.assignVehicle('tenant-100', 'v-101', 'drv-999'),
        throwsA(isA<Exception>()),
      );
    });

    test('unassignVehicle sends POST and returns true on 200', () async {
      final mockClient = MockClient((request) async {
        expect(request.method, 'POST');
        expect(request.url.path, '/api/v1/tenants/tenant-100/vehicles/v-101/unassign');

        return http.Response(
          jsonEncode({'success': true, 'data': {'status': 'AVAILABLE'}}),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);
      final success = await client.unassignVehicle('tenant-100', 'v-101');

      expect(success, isTrue);
    });
  });
}

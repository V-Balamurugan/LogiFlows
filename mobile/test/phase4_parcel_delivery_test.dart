import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/models/parcel_model.dart';
import 'package:logiflows_mobile/screens/delivery_screen.dart';
import 'package:logiflows_mobile/services/resource_api_client.dart';

void main() {
  group('Phase 4: Parcel & Delivery Lifecycle Models', () {
    test('ParcelModel parses valid JSON accurately', () {
      final json = {
        'id': 'p-1001',
        'tracking_number': 'PKG-20260924-DEMO-0001',
        'tenant_id': '00000000-0000-0000-0000-000000000001',
        'origin_branch_id': 'b-origin',
        'destination_branch_id': 'b-dest',
        'current_branch_id': 'b-origin',
        'sender_name': 'Alice Smith',
        'sender_phone': '+15550101',
        'sender_address': '100 Main St',
        'sender_city': 'New York',
        'recipient_name': 'Bob Jones',
        'recipient_phone': '+15550202',
        'recipient_address': '200 Market St',
        'recipient_city': 'Boston',
        'weight_kg': 4.5,
        'status': 'RECEIVED_AT_ORIGIN_BRANCH',
        'service_type': 'EXPRESS',
        'created_at': '2026-09-24T10:00:00Z',
        'updated_at': '2026-09-24T10:30:00Z',
      };

      final model = ParcelModel.fromJson(json);
      expect(model.id, 'p-1001');
      expect(model.trackingNumber, 'PKG-20260924-DEMO-0001');
      expect(model.weightKg, 4.5);
      expect(model.status, 'RECEIVED_AT_ORIGIN_BRANCH');
      expect(model.serviceType, 'EXPRESS');
      expect(model.senderCity, 'New York');
      expect(model.recipientCity, 'Boston');
    });

    test('DeliveryTaskModel parses valid JSON and supports serialization', () {
      final json = {
        'id': 'task-501',
        'tenant_id': '00000000-0000-0000-0000-000000000001',
        'parcel_id': 'p-1001',
        'parcel_tracking_number': 'PKG-20260924-DEMO-0001',
        'driver_id': 'drv-1',
        'driver_name': 'Dave Driver',
        'branch_id': 'b-dest',
        'status': 'ASSIGNED',
        'priority': 'URGENT',
        'recipient_name': 'Bob Jones',
        'recipient_phone': '+15550202',
        'recipient_address': '200 Market St',
        'created_at': '2026-09-24T11:00:00Z',
      };

      final model = DeliveryTaskModel.fromJson(json);
      expect(model.id, 'task-501');
      expect(model.priority, 'URGENT');
      expect(model.status, 'ASSIGNED');
      expect(model.parcelTrackingNumber, 'PKG-20260924-DEMO-0001');

      final serialized = model.toJson();
      expect(serialized['id'], 'task-501');
      expect(serialized['priority'], 'URGENT');
    });

    test('BranchTransferModel parses linehaul manifest details', () {
      final json = {
        'id': 'tx-901',
        'manifest_number': 'TRF-20260924-0001',
        'tenant_id': '00000000-0000-0000-0000-000000000001',
        'source_branch_id': 'b-origin',
        'destination_branch_id': 'b-dest',
        'status': 'DISPATCHED',
        'total_parcels': 12,
        'created_at': '2026-09-24T08:00:00Z',
      };

      final model = BranchTransferModel.fromJson(json);
      expect(model.id, 'tx-901');
      expect(model.manifestNumber, 'TRF-20260924-0001');
      expect(model.status, 'DISPATCHED');
      expect(model.totalParcels, 12);
    });

    test('PublicTrackingModel parses public event history without leaking sensitive IDs', () {
      final json = {
        'tracking_number': 'PKG-20260924-DEMO-0001',
        'status': 'IN_TRANSIT',
        'service_type': 'STANDARD',
        'origin_city': 'New York',
        'destination_city': 'Boston',
        'current_location': 'Philadelphia Sorting Hub',
        'events': [
          {
            'status': 'CREATED',
            'location': 'New York',
            'description': 'Shipping label generated',
            'timestamp': '2026-09-24T08:00:00Z',
          },
          {
            'status': 'IN_TRANSIT',
            'location': 'Philadelphia Sorting Hub',
            'description': 'In transit between hubs',
            'timestamp': '2026-09-24T12:00:00Z',
          }
        ]
      };

      final model = PublicTrackingModel.fromJson(json);
      expect(model.trackingNumber, 'PKG-20260924-DEMO-0001');
      expect(model.status, 'IN_TRANSIT');
      expect(model.events.length, 2);
      expect(model.events.first.status, 'CREATED');
      expect(model.events.last.location, 'Philadelphia Sorting Hub');
    });
  });

  group('Phase 4: ResourceApiClient Delivery Endpoints', () {
    const tenantId = '00000000-0000-0000-0000-000000000001';

    test('getDeliveryTasks calls API and parses list', () async {
      final mockHttp = MockClient((request) async {
        expect(request.url.path, contains('/tenants/$tenantId/deliveries'));
        return http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'delivery_tasks': [
                {
                  'id': 'task-1',
                  'tenant_id': tenantId,
                  'parcel_id': 'p-1',
                  'parcel_tracking_number': 'PKG-001',
                  'driver_id': 'drv-1',
                  'branch_id': 'b-1',
                  'status': 'ASSIGNED',
                  'priority': 'NORMAL',
                  'created_at': '2026-09-24T10:00:00Z',
                }
              ]
            }
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final storage = InMemorySecureTokenStorage();
      await storage.saveTokens(accessToken: 'mock_token', refreshToken: 'mock_ref');
      final client = ResourceApiClient(httpClient: mockHttp, tokenStorage: storage);

      final tasks = await client.getDeliveryTasks(tenantId);
      expect(tasks.length, 1);
      expect(tasks.first.parcelTrackingNumber, 'PKG-001');
      expect(tasks.first.status, 'ASSIGNED');
    });

    test('recordDeliveryAttempt sends correct payload', () async {
      final mockHttp = MockClient((request) async {
        expect(request.url.path, contains('/deliveries/task-1/attempts'));
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['outcome'], 'FAILED');
        expect(body['reason'], 'CUSTOMER_UNAVAILABLE');

        return http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'id': 'att-1',
              'task_id': 'task-1',
              'attempt_number': 1,
              'outcome': 'FAILED',
              'reason': 'CUSTOMER_UNAVAILABLE',
              'attempted_at': '2026-09-24T12:00:00Z',
            }
          }),
          201,
          headers: {'content-type': 'application/json'},
        );
      });

      final storage = InMemorySecureTokenStorage();
      await storage.saveTokens(accessToken: 'mock_token', refreshToken: 'mock_ref');
      final client = ResourceApiClient(httpClient: mockHttp, tokenStorage: storage);

      final attempt = await client.recordDeliveryAttempt(
        tenantId,
        'task-1',
        outcome: 'FAILED',
        reason: 'CUSTOMER_UNAVAILABLE',
        notes: 'Gate was locked',
      );

      expect(attempt.id, 'att-1');
      expect(attempt.outcome, 'FAILED');
    });

    test('submitDeliveryProof completes task with status 200', () async {
      final mockHttp = MockClient((request) async {
        expect(request.url.path, contains('/deliveries/task-1/proof'));
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['proof_type'], 'SIGNATURE');
        expect(body['recipient_name'], 'Jane Doe');

        return http.Response(
          jsonEncode({
            'success': true,
            'message': 'Delivery proof verified, parcel marked DELIVERED',
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final storage = InMemorySecureTokenStorage();
      await storage.saveTokens(accessToken: 'mock_token', refreshToken: 'mock_ref');
      final client = ResourceApiClient(httpClient: mockHttp, tokenStorage: storage);

      final success = await client.submitDeliveryProof(
        tenantId,
        'task-1',
        proofType: 'SIGNATURE',
        recipientName: 'Jane Doe',
        signatureData: 'SIG_DATA_BLOB',
      );

      expect(success, isTrue);
    });

    test('verifyParcelScan checks QR code payload', () async {
      final mockHttp = MockClient((request) async {
        expect(request.url.path, contains('/parcels/scan'));
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['qr_payload'], 'PKG-20260924-DEMO-0001');

        return http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'id': 'p-1',
              'tracking_number': 'PKG-20260924-DEMO-0001',
              'status': 'IN_TRANSIT',
              'weight_kg': 2.5,
            }
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final storage = InMemorySecureTokenStorage();
      await storage.saveTokens(accessToken: 'mock_token', refreshToken: 'mock_ref');
      final client = ResourceApiClient(httpClient: mockHttp, tokenStorage: storage);

      final res = await client.verifyParcelScan(tenantId, 'PKG-20260924-DEMO-0001');
      expect(res['tracking_number'], 'PKG-20260924-DEMO-0001');
      expect(res['status'], 'IN_TRANSIT');
    });

    test('getPublicTracking returns public sanitized tracking model', () async {
      final mockHttp = MockClient((request) async {
        expect(request.url.path, '/api/v1/tracking/PKG-PUBLIC-99');
        return http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'tracking_number': 'PKG-PUBLIC-99',
              'status': 'DELIVERED',
              'service_type': 'STANDARD',
              'events': []
            }
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final client = ResourceApiClient(httpClient: mockHttp);
      final tracking = await client.getPublicTracking('PKG-PUBLIC-99');
      expect(tracking.trackingNumber, 'PKG-PUBLIC-99');
      expect(tracking.status, 'DELIVERED');
    });
  });

  group('Phase 4: ParcelDeliveryScreen Widget Tests', () {
    const tenantId = '00000000-0000-0000-0000-000000000001';

    testWidgets('renders delivery screen tabs and displays task list', (WidgetTester tester) async {
      final mockHttp = MockClient((request) async {
        if (request.url.path.contains('/deliveries')) {
          return http.Response(
            jsonEncode({
              'success': true,
              'data': {
                'delivery_tasks': [
                  {
                    'id': 'task-100',
                    'tenant_id': tenantId,
                    'parcel_id': 'p-100',
                    'parcel_tracking_number': 'PKG-URGENT-001',
                    'driver_id': 'drv-1',
                    'branch_id': 'b-1',
                    'status': 'ASSIGNED',
                    'priority': 'URGENT',
                    'recipient_name': 'Carlos Courier',
                    'recipient_phone': '+15559999',
                    'recipient_address': '456 Delivery Lane',
                    'created_at': '2026-09-24T10:00:00Z',
                  }
                ]
              }
            }),
            200,
            headers: {'content-type': 'application/json'},
          );
        } else if (request.url.path.contains('/transfers')) {
          return http.Response(
            jsonEncode({
              'success': true,
              'data': {'transfers': []}
            }),
            200,
            headers: {'content-type': 'application/json'},
          );
        }
        return http.Response('{"success": true}', 200);
      });

      final storage = InMemorySecureTokenStorage();
      await storage.saveTokens(accessToken: 'mock_token', refreshToken: 'mock_ref');
      final client = ResourceApiClient(httpClient: mockHttp, tokenStorage: storage);

      await tester.pumpWidget(
        MaterialApp(
          home: ParcelDeliveryScreen(
            tenantId: tenantId,
            apiClient: client,
          ),
        ),
      );

      // Settle async fetch
      await tester.pumpAndSettle();

      expect(find.text('LogiFlows Dispatch Pro'), findsOneWidget);
      expect(find.text('Tasks (1)'), findsOneWidget);
      expect(find.text('Scan & Custody'), findsOneWidget);
      expect(find.text('Transfers (0)'), findsOneWidget);

      expect(find.text('PKG-URGENT-001'), findsOneWidget);
      expect(find.text('URGENT'), findsOneWidget);
      expect(find.text('Carlos Courier'), findsOneWidget);
      expect(find.text('+15559999'), findsOneWidget);
      expect(find.text('Attempt'), findsOneWidget);
      expect(find.text('Complete (POD)'), findsOneWidget);
    });

    testWidgets('navigates to Scan & Custody tab and renders barcode scan interface', (WidgetTester tester) async {
      final mockHttp = MockClient((request) async {
        return http.Response(
          jsonEncode({
            'success': true,
            'data': {'delivery_tasks': [], 'transfers': []}
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final storage = InMemorySecureTokenStorage();
      await storage.saveTokens(accessToken: 'mock_token', refreshToken: 'mock_ref');
      final client = ResourceApiClient(httpClient: mockHttp, tokenStorage: storage);

      await tester.pumpWidget(
        MaterialApp(
          home: ParcelDeliveryScreen(
            tenantId: tenantId,
            apiClient: client,
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Tap on Scan & Custody tab
      await tester.tap(find.text('Scan & Custody'));
      await tester.pumpAndSettle();

      expect(find.text('Digital Custody Scanner'), findsOneWidget);
      expect(find.text('Verify Parcel Code'), findsOneWidget);
    });
  });
}

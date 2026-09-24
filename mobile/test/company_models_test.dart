import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/models/resource_models.dart';
import 'package:logiflows_mobile/services/resource_api_client.dart';

void main() {
  group('Company / Tenant Mobile Unit Tests (Phase 2)', () {
    late TokenStorage tokenStorage;

    setUp(() async {
      tokenStorage = InMemorySecureTokenStorage();
      await tokenStorage.clearTokens();
      await tokenStorage.saveTokens(accessToken: 'mock.admin.jwt');
    });

    test('CompanyModel deserializes correctly from complete JSON payload', () {
      final json = {
        'id': 'cmp-001',
        'name': 'Apex Express Logistics',
        'slug': 'apex-express',
        'status': 'ACTIVE',
        'contact_email': 'ops@apexexpress.in',
        'created_at': '2026-09-22T10:00:00Z',
      };

      final company = CompanyModel.fromJson(json);

      expect(company.id, 'cmp-001');
      expect(company.name, 'Apex Express Logistics');
      expect(company.slug, 'apex-express');
      expect(company.status, 'ACTIVE');
      expect(company.contactEmail, 'ops@apexexpress.in');
      expect(company.createdAt, '2026-09-22T10:00:00Z');
    });

    test('CompanyModel serializes correctly to JSON map', () {
      final company = CompanyModel(
        id: 'cmp-002',
        name: 'V-Bala Logistics',
        slug: 'vbala-logistics',
        status: 'ACTIVE',
        contactEmail: 'contact@vbala.com',
      );

      final json = company.toJson();

      expect(json['id'], 'cmp-002');
      expect(json['name'], 'V-Bala Logistics');
      expect(json['slug'], 'vbala-logistics');
      expect(json['status'], 'ACTIVE');
      expect(json['contact_email'], 'contact@vbala.com');
    });

    test('ResourceApiClient.getCompanyDetails fetches and deserializes company payload', () async {
      final mockClient = MockClient((request) async {
        expect(request.method, 'GET');
        expect(request.url.path, '/api/v1/tenants/cmp-001');
        expect(request.headers['Authorization'], 'Bearer mock.admin.jwt');

        final responseBody = jsonEncode({
          'success': true,
          'data': {
            'id': 'cmp-001',
            'name': 'Apex Express Logistics',
            'slug': 'apex-express',
            'status': 'ACTIVE',
            'contact_email': 'ops@apexexpress.in',
            'created_at': '2026-09-22T10:00:00Z',
          }
        });

        return http.Response(responseBody, 200, headers: {'content-type': 'application/json'});
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);
      final company = await client.getCompanyDetails('cmp-001');

      expect(company.id, 'cmp-001');
      expect(company.name, 'Apex Express Logistics');
      expect(company.status, 'ACTIVE');
    });

    test('ResourceApiClient.getCompanyDetails throws Exception on 404', () async {
      final mockClient = MockClient((request) async {
        return http.Response(
          jsonEncode({'success': false, 'error': {'code': 'NOT_FOUND', 'message': 'Company not found'}}),
          404,
          headers: {'content-type': 'application/json'},
        );
      });

      final client = ResourceApiClient(httpClient: mockClient, tokenStorage: tokenStorage);

      expect(
        () => client.getCompanyDetails('non-existent-id'),
        throwsA(isA<Exception>()),
      );
    });
  });
}

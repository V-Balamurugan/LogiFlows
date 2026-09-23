import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/api_config.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/services/auth_api_client.dart';

void main() {
  group('AuthApiClient Unit Tests (Phase 1 Identity & Authentication)', () {
    late TokenStorage tokenStorage;

    setUp(() async {
      tokenStorage = InMemorySecureTokenStorage();
      await tokenStorage.clearTokens();
    });

    final mockValidAuthResponseJson = jsonEncode({
      'data': {
        'token': 'jwt.access.token.driver1',
        'expires_at': '2026-09-22T21:00:00Z',
        'refresh_token': 'rt.driver1.uuid',
        'refresh_token_expires_at': '2026-09-29T21:00:00Z',
        'user': {
          'id': 'usr-101',
          'email': 'driver1@logiflows.test',
          'full_name': 'Driver One',
          'is_active': true,
          'is_platform_admin': false,
          'email_verified': true,
        },
        'tenants': [
          {
            'id': 'ten-001',
            'name': 'Apex Logistics',
            'slug': 'apex-logistics',
            'role': 'DRIVER',
          }
        ],
      }
    });

    test('TC-P1-MOB-004: login succeeds and persists token pair into TokenStorage', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.toString(), ApiConfig.loginEndpoint);
        expect(request.method, 'POST');
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['email'], 'driver1@logiflows.test');
        expect(body['password'], 'Password123!');

        return http.Response(mockValidAuthResponseJson, 200, headers: {'content-type': 'application/json'});
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      final response = await authClient.login(
        email: '  Driver1@LogiFlows.Test  ', // Tests normalization
        password: 'Password123!',
      );

      expect(response.token, 'jwt.access.token.driver1');
      expect(response.refreshToken, 'rt.driver1.uuid');
      expect(response.user.email, 'driver1@logiflows.test');
      expect(response.tenants.first.name, 'Apex Logistics');

      // Verify token persistence
      expect(await tokenStorage.getAccessToken(), 'jwt.access.token.driver1');
      expect(await tokenStorage.getRefreshToken(), 'rt.driver1.uuid');
      expect(await tokenStorage.hasValidToken(), isTrue);
    });

    test('TC-P1-MOB-005: login fails and throws formatted exception message on 401 Unauthorized', () async {
      final mockClient = MockClient((request) async {
        final errorPayload = jsonEncode({
          'error': {
            'code': 'INVALID_CREDENTIALS',
            'message': 'Invalid email or password',
          }
        });
        return http.Response(errorPayload, 401, headers: {'content-type': 'application/json'});
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await expectLater(
        () => authClient.login(email: 'wrong@logiflows.test', password: 'WrongPassword'),
        throwsA(predicate((e) => e is Exception && e.toString().contains('Invalid email or password'))),
      );
      expect(await tokenStorage.hasValidToken(), isFalse);
    });

    test('TC-P1-MOB-006: register creates organization and driver account, persisting tokens', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.toString(), ApiConfig.registerEndpoint);
        expect(request.method, 'POST');
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['email'], 'newdriver@logiflows.test');
        expect(body['full_name'], 'Ravi Sharma');
        expect(body['company_name'], 'Bharat Freight Ltd');
        expect(body['phone_number'], '+919988776655');

        return http.Response(mockValidAuthResponseJson, 201, headers: {'content-type': 'application/json'});
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      final response = await authClient.register(
        email: 'NewDriver@LogiFlows.Test',
        password: 'Password123!',
        fullName: 'Ravi Sharma',
        companyName: 'Bharat Freight Ltd',
        phoneNumber: '+919988776655',
      );

      expect(response.token, 'jwt.access.token.driver1');
      expect(await tokenStorage.hasValidToken(), isTrue);
    });

    test('TC-P1-MOB-007: getCurrentUser succeeds when valid access token is present in storage', () async {
      await tokenStorage.saveTokens(accessToken: 'stored.access.token');

      final mockClient = MockClient((request) async {
        expect(request.url.toString(), ApiConfig.meEndpoint);
        expect(request.headers['Authorization'], 'Bearer stored.access.token');
        return http.Response(
          jsonEncode({
            'data': {
              'user': {
                'id': 'usr-me-1',
                'email': 'driver@me.test',
                'full_name': 'Profile User',
                'is_active': true,
                'is_platform_admin': false,
                'email_verified': true,
              }
            }
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);
      final user = await authClient.getCurrentUser();

      expect(user.id, 'usr-me-1');
      expect(user.email, 'driver@me.test');
      expect(user.fullName, 'Profile User');
    });

    test('getCurrentUser throws when no access token exists', () async {
      final mockClient = MockClient((request) async => http.Response('{}', 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await expectLater(
        () => authClient.getCurrentUser(),
        throwsA(predicate((e) => e is Exception && e.toString().contains('No access token found'))),
      );
    });

    test('TC-P1-MOB-008: refreshToken updates access and refresh tokens when valid', () async {
      await tokenStorage.saveTokens(
        accessToken: 'expired-access',
        refreshToken: 'valid-refresh-token',
      );

      final mockClient = MockClient((request) async {
        expect(request.url.toString(), ApiConfig.refreshEndpoint);
        final body = jsonDecode(request.body) as Map<String, dynamic>;
        expect(body['refresh_token'], 'valid-refresh-token');

        return http.Response(
          jsonEncode({
            'data': {
              'token': 'brand-new-access-token',
              'expires_at': '2026-09-22T21:15:00Z',
              'refresh_token': 'rotated-refresh-token',
              'refresh_token_expires_at': '2026-09-29T21:15:00Z',
              'user': {
                'id': 'usr-refreshed',
                'email': 'driver@refreshed.test',
                'full_name': 'Refreshed Driver',
                'is_active': true,
                'is_platform_admin': false,
                'email_verified': true,
              },
              'tenants': [],
            }
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);
      final refreshedResp = await authClient.refreshToken();

      expect(refreshedResp.token, 'brand-new-access-token');
      expect(await tokenStorage.getAccessToken(), 'brand-new-access-token');
      expect(await tokenStorage.getRefreshToken(), 'rotated-refresh-token');
    });

    test('refreshToken clears tokens and throws when refresh fails', () async {
      await tokenStorage.saveTokens(
        accessToken: 'access-1',
        refreshToken: 'revoked-refresh-token',
      );

      final mockClient = MockClient((request) async {
        return http.Response('{"error":{"message":"Invalid refresh token"}}', 401);
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await expectLater(
        () => authClient.refreshToken(),
        throwsA(predicate((e) => e is Exception && e.toString().contains('Session revoked'))),
      );
      expect(await tokenStorage.hasValidToken(), isFalse);
    });

    test('TC-P1-MOB-009: logout invalidates server session and clears local storage safely', () async {
      await tokenStorage.saveTokens(
        accessToken: 'active-access-token',
        refreshToken: 'active-refresh-token',
      );

      bool logoutEndpointHit = false;
      final mockClient = MockClient((request) async {
        if (request.url.toString() == ApiConfig.logoutEndpoint) {
          logoutEndpointHit = true;
          expect(request.headers['Authorization'], 'Bearer active-access-token');
          final body = jsonDecode(request.body) as Map<String, dynamic>;
          expect(body['refresh_token'], 'active-refresh-token');
          return http.Response(jsonEncode({'data': {'status': 'logged_out'}}), 200);
        }
        return http.Response('{}', 404);
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);
      await authClient.logout();

      expect(logoutEndpointHit, isTrue);
      expect(await tokenStorage.hasValidToken(), isFalse);
      expect(await tokenStorage.getAccessToken(), isNull);
      expect(await tokenStorage.getRefreshToken(), isNull);
    });

    test('logout clears tokens even if server endpoint responds with network error', () async {
      await tokenStorage.saveTokens(accessToken: 'token-err', refreshToken: 'rt-err');

      final mockClient = MockClient((request) async {
        throw Exception('Network unreachable / Connection refused');
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);
      await authClient.logout();

      // Ensure storage is still wiped cleanly
      expect(await tokenStorage.hasValidToken(), isFalse);
      expect(await tokenStorage.getAccessToken(), isNull);
    });

    test('TC-P1-MOB-015: login catches network connection failure and throws user-friendly exception', () async {
      final mockClient = MockClient((request) async {
        throw http.ClientException('Failed to fetch', request.url);
      });

      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await expectLater(
        () => authClient.login(email: 'admin1@gmail.com', password: 'Password123!'),
        throwsA(predicate((e) =>
            e is Exception &&
            e.toString().contains('Unable to connect to LogiFlows API'))),
      );
      expect(await tokenStorage.hasValidToken(), isFalse);
    });
  });
}


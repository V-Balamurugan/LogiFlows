import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/screens/register_screen.dart';
import 'package:logiflows_mobile/services/auth_api_client.dart';

void main() {
  group('RegisterScreen Widget Tests (Phase 1 Identity & Multi-Tenancy)', () {
    late TokenStorage tokenStorage;

    setUp(() async {
      tokenStorage = InMemorySecureTokenStorage();
      await tokenStorage.clearTokens();
    });

    testWidgets('TC-P1-MOB-013: renders registration fields and validates required input fields', (WidgetTester tester) async {
      final mockClient = MockClient((request) async => http.Response('{}', 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: RegisterScreen(
            authClient: authClient,
            onRegisterSuccess: () {},
          ),
        ),
      );

      expect(find.text('Register Logistics Company'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Full Name'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Company / Fleet Name'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Corporate Email'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Password (min 8 chars, mixed case, symbols)'), findsOneWidget);
      expect(find.widgetWithText(ElevatedButton, 'Create Company & Account'), findsOneWidget);

      // Tap submit with empty fields
      await tester.tap(find.widgetWithText(ElevatedButton, 'Create Company & Account'));
      await tester.pumpAndSettle();

      expect(find.text('Full name is required'), findsOneWidget);
      expect(find.text('Company name is required'), findsOneWidget);
      expect(find.text('Valid email is required'), findsOneWidget);
      expect(find.text('Minimum 8 characters'), findsOneWidget);
    });

    testWidgets('successfully registers company, pops screen, and invokes onRegisterSuccess', (WidgetTester tester) async {
      final mockSuccessResponse = jsonEncode({
        'data': {
          'token': 'jwt.token.registered',
          'expires_at': '2026-09-22T21:00:00Z',
          'refresh_token': 'refresh.uuid',
          'user': {
            'id': 'usr-reg-1',
            'email': 'reg@logiflows.test',
            'full_name': 'Registered User',
            'is_active': true,
            'is_platform_admin': false,
            'email_verified': true,
          },
          'tenants': [
            {
              'id': 'ten-1',
              'name': 'New Fleet Logistics',
              'slug': 'new-fleet-logistics',
              'role': 'TENANT_ADMIN',
            }
          ],
        }
      });

      final mockClient = MockClient((request) async => http.Response(mockSuccessResponse, 201));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      bool registerSuccessTriggered = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Navigator(
            onGenerateRoute: (_) => MaterialPageRoute(
              builder: (ctx) => RegisterScreen(
                authClient: authClient,
                onRegisterSuccess: () {
                  registerSuccessTriggered = true;
                },
              ),
            ),
          ),
        ),
      );

      await tester.enterText(find.widgetWithText(TextFormField, 'Full Name'), 'Registered User');
      await tester.enterText(find.widgetWithText(TextFormField, 'Company / Fleet Name'), 'New Fleet Logistics');
      await tester.enterText(find.widgetWithText(TextFormField, 'Corporate Email'), 'reg@logiflows.test');
      await tester.enterText(find.widgetWithText(TextFormField, 'Password (min 8 chars, mixed case, symbols)'), 'SecurePassword123!');

      await tester.tap(find.widgetWithText(ElevatedButton, 'Create Company & Account'));
      await tester.pumpAndSettle();

      expect(registerSuccessTriggered, isTrue);
    });

    testWidgets('displays error message when registration endpoint returns 409 Conflict', (WidgetTester tester) async {
      final mockConflictResponse = jsonEncode({
        'error': {
          'code': 'EMAIL_ALREADY_EXISTS',
          'message': 'A company or user with this email address already exists',
        }
      });

      final mockClient = MockClient((request) async => http.Response(mockConflictResponse, 409));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: RegisterScreen(
            authClient: authClient,
            onRegisterSuccess: () {},
          ),
        ),
      );

      await tester.enterText(find.widgetWithText(TextFormField, 'Full Name'), 'Registered User');
      await tester.enterText(find.widgetWithText(TextFormField, 'Company / Fleet Name'), 'Existing Fleet');
      await tester.enterText(find.widgetWithText(TextFormField, 'Corporate Email'), 'existing@logiflows.test');
      await tester.enterText(find.widgetWithText(TextFormField, 'Password (min 8 chars, mixed case, symbols)'), 'SecurePassword123!');

      await tester.tap(find.widgetWithText(ElevatedButton, 'Create Company & Account'));
      await tester.pumpAndSettle();

      expect(find.text('A company or user with this email address already exists'), findsOneWidget);
    });
  });
}

import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/screens/login_screen.dart';
import 'package:logiflows_mobile/screens/register_screen.dart';
import 'package:logiflows_mobile/services/auth_api_client.dart';

void main() {
  group('LoginScreen Widget Tests (Phase 1 Identity & Authentication)', () {
    late TokenStorage tokenStorage;

    setUp(() async {
      tokenStorage = InMemorySecureTokenStorage();
      await tokenStorage.clearTokens();
    });

    testWidgets('TC-P1-MOB-010: renders branding, input fields, sign in button, and register link', (WidgetTester tester) async {
      final mockClient = MockClient((request) async => http.Response('{}', 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: LoginScreen(
            authClient: authClient,
            onLoginSuccess: () {},
          ),
        ),
      );

      expect(find.text('LogiFlows Driver Pro'), findsOneWidget);
      expect(find.text('Sign in to access assigned shipments & routes'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Corporate Email'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Password'), findsOneWidget);
      expect(find.widgetWithText(ElevatedButton, 'Sign In'), findsOneWidget);
      expect(find.text('Register Company'), findsOneWidget);
    });

    testWidgets('TC-P1-MOB-011: validates required fields, email format, and password length', (WidgetTester tester) async {
      final mockClient = MockClient((request) async => http.Response('{}', 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: LoginScreen(
            authClient: authClient,
            onLoginSuccess: () {},
          ),
        ),
      );

      // Tap Sign In with empty fields
      await tester.tap(find.widgetWithText(ElevatedButton, 'Sign In'));
      await tester.pumpAndSettle();

      expect(find.text('Email is required'), findsOneWidget);
      expect(find.text('Password is required'), findsOneWidget);

      // Enter invalid email and short password
      await tester.enterText(find.widgetWithText(TextFormField, 'Corporate Email'), 'invalid-email');
      await tester.enterText(find.widgetWithText(TextFormField, 'Password'), 'short');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Sign In'));
      await tester.pumpAndSettle();

      expect(find.text('Enter a valid email address'), findsOneWidget);
      expect(find.text('Password must be at least 8 characters'), findsOneWidget);
    });

    testWidgets('toggles password visibility when eye suffix icon is clicked', (WidgetTester tester) async {
      final mockClient = MockClient((request) async => http.Response('{}', 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: LoginScreen(
            authClient: authClient,
            onLoginSuccess: () {},
          ),
        ),
      );

      expect(find.byIcon(Icons.visibility_off), findsOneWidget);
      expect(find.byIcon(Icons.visibility), findsNothing);

      // Tap visibility toggle
      await tester.tap(find.byIcon(Icons.visibility_off));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.visibility), findsOneWidget);
      expect(find.byIcon(Icons.visibility_off), findsNothing);
    });

    testWidgets('executes login and triggers onLoginSuccess callback on 200 response', (WidgetTester tester) async {
      final mockSuccessResponse = jsonEncode({
        'data': {
          'token': 'jwt.token.valid',
          'expires_at': '2026-09-22T21:00:00Z',
          'refresh_token': 'refresh.uuid',
          'user': {
            'id': 'usr-1',
            'email': 'driver@test.com',
            'full_name': 'Valid Driver',
            'is_active': true,
            'is_platform_admin': false,
            'email_verified': true,
          },
          'tenants': [],
        }
      });

      final mockClient = MockClient((request) async => http.Response(mockSuccessResponse, 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      bool loginSuccessTriggered = false;

      await tester.pumpWidget(
        MaterialApp(
          home: LoginScreen(
            authClient: authClient,
            onLoginSuccess: () {
              loginSuccessTriggered = true;
            },
          ),
        ),
      );

      await tester.enterText(find.widgetWithText(TextFormField, 'Corporate Email'), 'driver@test.com');
      await tester.enterText(find.widgetWithText(TextFormField, 'Password'), 'Password123!');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Sign In'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(loginSuccessTriggered, isTrue);
    });

    testWidgets('TC-P1-MOB-012: displays error banner when login authentication fails', (WidgetTester tester) async {
      final mockErrorResponse = jsonEncode({
        'error': {
          'code': 'INVALID_CREDENTIALS',
          'message': 'Invalid corporate email or password provided',
        }
      });

      final mockClient = MockClient((request) async => http.Response(mockErrorResponse, 401));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: LoginScreen(
            authClient: authClient,
            onLoginSuccess: () {},
          ),
        ),
      );

      await tester.enterText(find.widgetWithText(TextFormField, 'Corporate Email'), 'driver@test.com');
      await tester.enterText(find.widgetWithText(TextFormField, 'Password'), 'WrongPassword123!');
      await tester.tap(find.widgetWithText(ElevatedButton, 'Sign In'));
      await tester.pumpAndSettle();

      expect(find.text('Invalid corporate email or password provided'), findsOneWidget);
      expect(find.byIcon(Icons.error_outline), findsOneWidget);
    });

    testWidgets('navigates to RegisterScreen when Register Company link is tapped', (WidgetTester tester) async {
      final mockClient = MockClient((request) async => http.Response('{}', 200));
      final authClient = AuthApiClient(tokenStorage: tokenStorage, httpClient: mockClient);

      await tester.pumpWidget(
        MaterialApp(
          home: LoginScreen(
            authClient: authClient,
            onLoginSuccess: () {},
          ),
        ),
      );

      await tester.tap(find.text('Register Company'));
      await tester.pumpAndSettle();

      expect(find.byType(RegisterScreen), findsOneWidget);
      expect(find.text('Register Logistics Company'), findsOneWidget);
    });
  });
}

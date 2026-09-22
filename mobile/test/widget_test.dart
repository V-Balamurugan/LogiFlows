import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/core/token_storage.dart';
import 'package:logiflows_mobile/main.dart';
import 'package:logiflows_mobile/screens/login_screen.dart';
import 'package:logiflows_mobile/services/auth_api_client.dart';

void main() {
  group('Mobile Application Flow & Widget Tests (Phase 0 & Phase 1)', () {
    setUp(() async {
      await InMemorySecureTokenStorage().clearTokens();
    });

    testWidgets('TC-P0-MOB-005: DriverDashboardScreen renders title, probe action, and assignment feed', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: DriverDashboardScreen(
            onLogout: () {},
          ),
        ),
      );

      expect(find.text('LogiFlows Driver Pro'), findsOneWidget);
      expect(find.text('System Connectivity'), findsOneWidget);
      expect(find.text('Active Custody Assignments'), findsOneWidget);
      expect(find.text('TRK-2026-9041'), findsOneWidget);
      expect(find.text('TRK-2026-9042'), findsOneWidget);
    });

    testWidgets('TC-P1-MOB-010: LoginScreen renders LogiFlows brand, fields, and login action', (WidgetTester tester) async {
      final authClient = AuthApiClient();

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
      expect(find.byType(TextFormField), findsNWidgets(2));
      expect(find.widgetWithText(ElevatedButton, 'Sign In'), findsOneWidget);
      expect(find.text('Register Company'), findsOneWidget);
    });

    testWidgets('TC-P0-MOB-007: LogiFlowsApp displays loading indicator during auth check and resolves to LoginScreen', (WidgetTester tester) async {
      await tester.pumpWidget(const LogiFlowsApp());

      // While initial Future in initState is running, circular progress is displayed
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      // Settle async token check
      await tester.pumpAndSettle();

      expect(find.byType(LoginScreen), findsOneWidget);
      expect(find.text('LogiFlows Driver Pro'), findsOneWidget);
      expect(find.text('Sign in to access assigned shipments & routes'), findsOneWidget);
    });

    testWidgets('TC-P1-MOB-014: LogiFlowsApp routes directly to DriverDashboardScreen when pre-authenticated token exists', (WidgetTester tester) async {
      // Pre-seed storage with an existing valid session token
      await InMemorySecureTokenStorage().saveTokens(accessToken: 'pre-authenticated-token');

      await tester.pumpWidget(const LogiFlowsApp());
      await tester.pumpAndSettle();

      expect(find.byType(DriverDashboardScreen), findsOneWidget);
      expect(find.text('LogiFlows Driver Pro'), findsOneWidget);
      expect(find.text('Active Custody Assignments'), findsOneWidget);
      expect(find.byType(LoginScreen), findsNothing);
    });
  });
}

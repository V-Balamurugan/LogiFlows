import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/main.dart';

void main() {
  group('DriverDashboardScreen Tests (Phase 0 Foundation & Phase 1 Custody Feed)', () {
    testWidgets('TC-P0-MOB-005: renders header, connectivity card, and custody parcels', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: DriverDashboardScreen(
            onLogout: () {},
          ),
        ),
      );

      expect(find.text('LogiFlows Driver Pro'), findsOneWidget);
      expect(find.byIcon(Icons.local_shipping), findsOneWidget);
      expect(find.byIcon(Icons.logout), findsOneWidget);
      expect(find.text('System Connectivity'), findsOneWidget);
      expect(find.text('Session Active'), findsOneWidget);
      expect(find.text('Probe LogiFlows Network'), findsOneWidget);

      expect(find.text('Active Custody Assignments'), findsOneWidget);
      expect(find.text('TRK-2026-9041'), findsOneWidget);
      expect(find.text('North Hub -> Central Branch'), findsOneWidget);
      expect(find.text('In Transit'), findsOneWidget);
      expect(find.text('ETA: 14:30 IST'), findsOneWidget);

      expect(find.text('TRK-2026-9042'), findsOneWidget);
      expect(find.text('Central Branch -> Airport Station'), findsOneWidget);
      expect(find.text('Out for Delivery'), findsOneWidget);
      expect(find.text('ETA: 15:15 IST'), findsOneWidget);
    });

    testWidgets('TC-P0-MOB-006: probe action transitions connectivity status from probing to operational', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: DriverDashboardScreen(
            onLogout: () {},
          ),
        ),
      );

      expect(find.text('Session Active'), findsOneWidget);

      // Tap Probe Network button
      await tester.tap(find.text('Probe LogiFlows Network'));
      await tester.pump(); // Start async probe

      expect(find.text('Probing Network...'), findsOneWidget);

      // Advance clock past the probe delay (600ms)
      await tester.pump(const Duration(milliseconds: 650));

      expect(find.text('Backend Operational'), findsOneWidget);
    });

    testWidgets('invokes onLogout callback when sign out action button is tapped', (WidgetTester tester) async {
      bool logoutInvoked = false;

      await tester.pumpWidget(
        MaterialApp(
          home: DriverDashboardScreen(
            onLogout: () {
              logoutInvoked = true;
            },
          ),
        ),
      );

      await tester.tap(find.byIcon(Icons.logout));
      await tester.pumpAndSettle();

      expect(logoutInvoked, isTrue);
    });
  });
}

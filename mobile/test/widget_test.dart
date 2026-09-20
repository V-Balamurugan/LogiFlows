import 'package:flutter_test/flutter_test.dart';
import 'package:logiflows_mobile/main.dart';

void main() {
  testWidgets('LogiFlows Driver App renders title and initial state', (WidgetTester tester) async {
    await tester.pumpWidget(const LogiFlowsApp());

    expect(find.text('LogiFlows Driver Pro'), findsOneWidget);
    expect(find.text('System Connectivity'), findsOneWidget);
    expect(find.text('Active Custody Assignments'), findsOneWidget);
  });
}

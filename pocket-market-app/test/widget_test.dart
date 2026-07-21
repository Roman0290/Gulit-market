import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:pocket_market_app/main.dart';

void main() {
  testWidgets('app boots and shows a loading state while the session is restored', (tester) async {
    await tester.pumpWidget(const PocketMarketApp());

    // restoreSession() is still in flight (real network/prefs calls aren't
    // mocked here), so the auth gate should show its loading spinner rather
    // than crash.
    await tester.pump();
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}

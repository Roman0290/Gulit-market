class AppConfig {
  // 10.0.2.2 is the Android emulator's alias for the host machine's
  // localhost; web/desktop/iOS-simulator builds use localhost directly.
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080/api/v1',
  );

  static const String stripePublishableKey = String.fromEnvironment(
    'STRIPE_PUBLISHABLE_KEY',
    defaultValue: '',
  );
}

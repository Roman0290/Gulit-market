import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_stripe/flutter_stripe.dart';
import 'package:provider/provider.dart';

import 'core/api/api_client.dart';
import 'core/api/token_store.dart';
import 'core/config/app_config.dart';
import 'core/theme/app_theme.dart';
import 'core/widgets/home_shell.dart';
import 'features/addresses/data/addresses_repository.dart';
import 'features/auth/data/auth_repository.dart';
import 'features/auth/presentation/auth_provider.dart';
import 'features/auth/presentation/login_screen.dart';
import 'features/cart/data/cart_repository.dart';
import 'features/cart/presentation/cart_provider.dart';
import 'features/orders/data/orders_repository.dart';
import 'features/orders/presentation/orders_provider.dart';
import 'features/payments/data/payments_repository.dart';
import 'features/products/data/products_repository.dart';
import 'features/products/presentation/products_provider.dart';

bool get _stripeSupported => !kIsWeb && (Platform.isAndroid || Platform.isIOS);

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  if (_stripeSupported && AppConfig.stripePublishableKey.isNotEmpty) {
    Stripe.publishableKey = AppConfig.stripePublishableKey;
    await Stripe.instance.applySettings();
  }

  runApp(const PocketMarketApp());
}

class PocketMarketApp extends StatelessWidget {
  const PocketMarketApp({super.key});

  @override
  Widget build(BuildContext context) {
    final tokenStore = TokenStore();
    final apiClient = ApiClient(tokenStore: tokenStore);

    return MultiProvider(
      providers: [
        Provider(create: (_) => AddressesRepository(apiClient)),
        Provider(create: (_) => PaymentsRepository(apiClient)),
        ChangeNotifierProvider(
          create: (_) => AuthProvider(AuthRepository(client: apiClient, tokenStore: tokenStore))
            ..restoreSession(),
        ),
        ChangeNotifierProvider(create: (_) => ProductsProvider(ProductsRepository(apiClient))),
        ChangeNotifierProvider(create: (_) => CartProvider(CartRepository(apiClient))),
        ChangeNotifierProvider(create: (_) => OrdersProvider(OrdersRepository(apiClient))),
      ],
      child: MaterialApp(
        title: 'ጉሊት Market',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.light(),
        home: const _AuthGate(),
      ),
    );
  }
}

class _AuthGate extends StatelessWidget {
  const _AuthGate();

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();

    switch (auth.status) {
      case AuthStatus.unknown:
        return const Scaffold(body: Center(child: CircularProgressIndicator()));
      case AuthStatus.unauthenticated:
        return const LoginScreen();
      case AuthStatus.authenticated:
        return const HomeShell();
    }
  }
}

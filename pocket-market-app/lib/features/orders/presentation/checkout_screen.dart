import 'dart:io' show Platform;

import 'package:collection/collection.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_stripe/flutter_stripe.dart' hide Address, Card;
import 'package:provider/provider.dart';

import '../../../core/api/api_exception.dart';
import '../../addresses/data/addresses_repository.dart';
import '../../addresses/domain/address.dart';
import '../../cart/presentation/cart_provider.dart';
import '../../payments/data/payments_repository.dart';
import '../domain/order.dart';
import 'order_history_screen.dart';
import 'orders_provider.dart';

/// flutter_stripe only ships a working platform implementation for
/// Android/iOS - on web/desktop we skip the card field and payment
/// confirmation step so the rest of the flow (order creation, cart
/// clearing) can still be exercised during development.
bool get _stripeSupported => !kIsWeb && (Platform.isAndroid || Platform.isIOS);

class CheckoutScreen extends StatefulWidget {
  const CheckoutScreen({super.key});

  @override
  State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  List<Address> _addresses = [];
  String? _selectedAddressId;
  String _paymentMethod = 'cod';
  bool _loadingAddresses = true;
  bool _placingOrder = false;
  String? _statusMessage;

  @override
  void initState() {
    super.initState();
    _loadAddresses();
  }

  Future<void> _loadAddresses() async {
    setState(() => _loadingAddresses = true);
    try {
      final addresses = await context.read<AddressesRepository>().list();
      setState(() {
        _addresses = addresses;
        _selectedAddressId = addresses.where((a) => a.isDefault).firstOrNull?.id ??
            addresses.firstOrNull?.id;
      });
    } catch (_) {
      // Leave the list empty - the "add address" prompt below covers it.
    } finally {
      setState(() => _loadingAddresses = false);
    }
  }

  Future<void> _addAddress() async {
    final result = await showModalBottomSheet<Address>(
      context: context,
      isScrollControlled: true,
      builder: (_) => const _AddAddressSheet(),
    );
    if (result != null) {
      setState(() {
        _addresses = [..._addresses, result];
        _selectedAddressId = result.id;
      });
    }
  }

  Future<void> _placeOrder() async {
    if (_selectedAddressId == null) return;

    setState(() {
      _placingOrder = true;
      _statusMessage = 'Creating your order...';
    });

    try {
      final createdOrders = await context.read<OrdersProvider>().checkout(
            addressId: _selectedAddressId!,
            paymentMethod: _paymentMethod,
          );
      if (!mounted) return;

      if (_paymentMethod == 'card' && _stripeSupported) {
        final payments = context.read<PaymentsRepository>();
        for (var i = 0; i < createdOrders.length; i++) {
          setState(() => _statusMessage =
              createdOrders.length > 1 ? 'Processing payment ${i + 1} of ${createdOrders.length}...' : 'Processing payment...');
          final intent = await payments.createIntent(createdOrders[i].id);
          await Stripe.instance.confirmPayment(
            paymentIntentClientSecret: intent.clientSecret,
            data: const PaymentMethodParams.card(paymentMethodData: PaymentMethodData()),
          );
        }
      }

      if (!mounted) return;
      context.read<CartProvider>().reset();
      _showSuccess(createdOrders);
    } on ApiException catch (e) {
      _showError(e.message);
    } on StripeException catch (e) {
      _showError(e.error.localizedMessage ?? 'Payment failed');
    } catch (e) {
      _showError('Something went wrong: $e');
    } finally {
      if (mounted) setState(() => _placingOrder = false);
    }
  }

  void _showSuccess(List<Order> orders) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.check_circle, color: Colors.green, size: 40),
        title: Text(orders.length > 1 ? 'Orders placed!' : 'Order placed!'),
        content: Text(
          orders.length > 1
              ? 'Your cart was split into ${orders.length} orders, one per vendor.'
              : 'Your order has been sent to the vendor.',
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).popUntil((route) => route.isFirst);
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const OrderHistoryScreen()),
              );
            },
            child: const Text('View orders'),
          ),
        ],
      ),
    );
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    final cart = context.watch<CartProvider>().cart;

    return Scaffold(
      appBar: AppBar(title: const Text('Checkout')),
      body: _loadingAddresses
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text('Delivery address', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                if (_addresses.isEmpty)
                  const Text('No saved addresses yet.')
                else
                  ..._addresses.map((a) => RadioListTile<String>(
                        value: a.id,
                        groupValue: _selectedAddressId,
                        onChanged: (v) => setState(() => _selectedAddressId = v),
                        title: Text(a.displayLabel),
                        subtitle: Text('${a.line1}, ${a.city}'),
                        contentPadding: EdgeInsets.zero,
                      )),
                OutlinedButton.icon(
                  onPressed: _addAddress,
                  icon: const Icon(Icons.add),
                  label: const Text('Add new address'),
                ),
                const SizedBox(height: 24),
                Text('Payment method', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                RadioListTile<String>(
                  value: 'cod',
                  groupValue: _paymentMethod,
                  onChanged: (v) => setState(() => _paymentMethod = v!),
                  title: const Text('Cash on delivery'),
                  contentPadding: EdgeInsets.zero,
                ),
                RadioListTile<String>(
                  value: 'card',
                  groupValue: _paymentMethod,
                  onChanged: (v) => setState(() => _paymentMethod = v!),
                  title: const Text('Card'),
                  contentPadding: EdgeInsets.zero,
                ),
                if (_paymentMethod == 'card') ...[
                  const SizedBox(height: 8),
                  if (_stripeSupported)
                    CardField(onCardChanged: (_) {})
                  else
                    Card(
                      color: Theme.of(context).colorScheme.surfaceContainerHighest,
                      child: const Padding(
                        padding: EdgeInsets.all(12),
                        child: Text(
                          'Card payment requires the Android or iOS app - this platform will create the '
                          'order with payment left pending.',
                          style: TextStyle(fontSize: 12),
                        ),
                      ),
                    ),
                ],
                const SizedBox(height: 24),
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text('Subtotal'),
                        Text(cart.subtotal.toStringAsFixed(2)),
                      ],
                    ),
                  ),
                ),
              ],
            ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: ElevatedButton(
            onPressed: (_selectedAddressId == null || _placingOrder) ? null : _placeOrder,
            child: _placingOrder
                ? Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const SizedBox(
                        height: 18,
                        width: 18,
                        child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                      ),
                      const SizedBox(width: 12),
                      Text(_statusMessage ?? 'Placing order...'),
                    ],
                  )
                : const Text('Place order'),
          ),
        ),
      ),
    );
  }
}

class _AddAddressSheet extends StatefulWidget {
  const _AddAddressSheet();

  @override
  State<_AddAddressSheet> createState() => _AddAddressSheetState();
}

class _AddAddressSheetState extends State<_AddAddressSheet> {
  final _formKey = GlobalKey<FormState>();
  final _labelController = TextEditingController();
  final _line1Controller = TextEditingController();
  final _cityController = TextEditingController();
  bool _saving = false;

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);
    try {
      final address = await context.read<AddressesRepository>().create(
            label: _labelController.text.trim(),
            line1: _line1Controller.text.trim(),
            city: _cityController.text.trim(),
            isDefault: true,
          );
      if (mounted) Navigator.of(context).pop(address);
    } on ApiException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('Add address', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 16),
            TextFormField(
              controller: _labelController,
              decoration: const InputDecoration(labelText: 'Label (e.g. Home)'),
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _line1Controller,
              decoration: const InputDecoration(labelText: 'Street address'),
              validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _cityController,
              decoration: const InputDecoration(labelText: 'City'),
              validator: (v) => (v == null || v.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: _saving ? null : _save,
              child: _saving
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Save address'),
            ),
          ],
        ),
      ),
    );
  }
}

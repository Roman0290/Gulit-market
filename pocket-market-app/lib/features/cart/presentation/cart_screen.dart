import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/widgets/async_state_view.dart';
import '../../orders/presentation/checkout_screen.dart';
import '../domain/cart.dart';
import 'cart_provider.dart';

class CartScreen extends StatefulWidget {
  const CartScreen({super.key});

  @override
  State<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends State<CartScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<CartProvider>().load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<CartProvider>();
    final cart = provider.cart;

    return Scaffold(
      appBar: AppBar(title: const Text('Cart')),
      body: AsyncStateView(
        isLoading: provider.isLoading,
        error: provider.error,
        isEmpty: cart.items.isEmpty,
        emptyMessage: 'Your cart is empty',
        onRetry: provider.load,
        builder: (context) => RefreshIndicator(
          onRefresh: provider.load,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (cart.vendorCount > 1)
                Card(
                  color: Theme.of(context).colorScheme.secondaryContainer,
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Row(
                      children: [
                        const Icon(Icons.info_outline, size: 18),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'Items from ${cart.vendorCount} vendors will be placed as separate orders.',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              const SizedBox(height: 8),
              ...cart.items.map((item) => _CartItemTile(item: item)),
            ],
          ),
        ),
      ),
      bottomNavigationBar: cart.items.isEmpty
          ? null
          : SafeArea(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Subtotal', style: Theme.of(context).textTheme.bodySmall),
                          Text(
                            cart.subtotal.toStringAsFixed(2),
                            style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
                          ),
                        ],
                      ),
                    ),
                    ElevatedButton(
                      onPressed: () => Navigator.of(context).push(
                        MaterialPageRoute(builder: (_) => const CheckoutScreen()),
                      ),
                      child: const Text('Checkout'),
                    ),
                  ],
                ),
              ),
            ),
    );
  }
}

class _CartItemTile extends StatelessWidget {
  final CartItem item;

  const _CartItemTile({required this.item});

  @override
  Widget build(BuildContext context) {
    final cart = context.read<CartProvider>();

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(item.productName, style: Theme.of(context).textTheme.titleSmall),
                  const SizedBox(height: 2),
                  Text('${item.unitPrice.toStringAsFixed(2)} / ${item.unit}'),
                ],
              ),
            ),
            IconButton.filledTonal(
              iconSize: 18,
              onPressed: item.quantity > 1
                  ? () => cart.updateQuantity(item.id, item.quantity - 1)
                  : () => cart.removeItem(item.id),
              icon: Icon(item.quantity > 1 ? Icons.remove : Icons.delete_outline),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Text('${item.quantity}'),
            ),
            IconButton.filledTonal(
              iconSize: 18,
              onPressed: () => cart.updateQuantity(item.id, item.quantity + 1),
              icon: const Icon(Icons.add),
            ),
            const SizedBox(width: 8),
            SizedBox(
              width: 64,
              child: Text(
                item.lineTotal.toStringAsFixed(2),
                textAlign: TextAlign.end,
                style: Theme.of(context).textTheme.titleSmall,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

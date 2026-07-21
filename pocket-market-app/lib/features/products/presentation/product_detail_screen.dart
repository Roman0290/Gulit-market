import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../cart/presentation/cart_provider.dart';
import '../domain/product.dart';

class ProductDetailScreen extends StatefulWidget {
  final Product product;

  const ProductDetailScreen({super.key, required this.product});

  @override
  State<ProductDetailScreen> createState() => _ProductDetailScreenState();
}

class _ProductDetailScreenState extends State<ProductDetailScreen> {
  int _quantity = 1;
  bool _adding = false;

  Future<void> _addToCart() async {
    setState(() => _adding = true);
    final cart = context.read<CartProvider>();
    final ok = await cart.addItem(widget.product.id, _quantity);
    if (mounted) {
      setState(() => _adding = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(ok ? 'Added to cart' : (cart.error ?? 'Could not add to cart'))),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final product = widget.product;
    final outOfStock = product.stockQuantity <= 0;

    return Scaffold(
      appBar: AppBar(title: Text(product.name)),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AspectRatio(
              aspectRatio: 1.3,
              child: Container(
                color: Theme.of(context).colorScheme.surfaceContainerHighest,
                child: product.imageUrl.isEmpty
                    ? Icon(Icons.eco, size: 64, color: Theme.of(context).colorScheme.primary)
                    : Image.network(
                        product.imageUrl,
                        fit: BoxFit.cover,
                        errorBuilder: (_, __, ___) => Icon(
                          Icons.eco,
                          size: 64,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(product.name, style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 4),
                  Text(
                    '${product.price.toStringAsFixed(2)} / ${product.unit}',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          color: Theme.of(context).colorScheme.primary,
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                  const SizedBox(height: 12),
                  if (product.description.isNotEmpty) Text(product.description),
                  const SizedBox(height: 8),
                  Text(
                    outOfStock ? 'Out of stock' : '${product.stockQuantity} ${product.unit} available',
                    style: TextStyle(
                      color: outOfStock ? Theme.of(context).colorScheme.error : Colors.grey.shade600,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              if (!outOfStock) ...[
                IconButton.filledTonal(
                  onPressed: _quantity > 1 ? () => setState(() => _quantity--) : null,
                  icon: const Icon(Icons.remove),
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Text('$_quantity', style: Theme.of(context).textTheme.titleMedium),
                ),
                IconButton.filledTonal(
                  onPressed: _quantity < product.stockQuantity ? () => setState(() => _quantity++) : null,
                  icon: const Icon(Icons.add),
                ),
                const SizedBox(width: 12),
              ],
              Expanded(
                child: ElevatedButton(
                  onPressed: (outOfStock || _adding) ? null : _addToCart,
                  child: _adding
                      ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                      : Text(outOfStock ? 'Out of stock' : 'Add to cart'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

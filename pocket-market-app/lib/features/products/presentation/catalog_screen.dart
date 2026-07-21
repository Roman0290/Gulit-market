import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/widgets/async_state_view.dart';
import '../domain/product.dart';
import 'product_detail_screen.dart';
import 'products_provider.dart';

class CatalogScreen extends StatefulWidget {
  const CatalogScreen({super.key});

  @override
  State<CatalogScreen> createState() => _CatalogScreenState();
}

class _CatalogScreenState extends State<CatalogScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<ProductsProvider>().loadInitial();
    });
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<ProductsProvider>();

    return Scaffold(
      appBar: AppBar(title: const Text('Gulit Market')),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
            child: TextField(
              decoration: const InputDecoration(
                hintText: 'Search vegetables, grains, oils...',
                prefixIcon: Icon(Icons.search),
              ),
              onChanged: provider.search,
            ),
          ),
          if (provider.categories.isNotEmpty)
            SizedBox(
              height: 40,
              child: ListView.separated(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: provider.categories.length,
                separatorBuilder: (_, __) => const SizedBox(width: 8),
                itemBuilder: (context, i) {
                  final category = provider.categories[i];
                  final selected = provider.selectedCategoryId == category.id;
                  return ChoiceChip(
                    label: Text(category.name),
                    selected: selected,
                    onSelected: (_) => provider.selectCategory(category.id),
                  );
                },
              ),
            ),
          const SizedBox(height: 8),
          Expanded(
            child: AsyncStateView(
              isLoading: provider.isLoading,
              error: provider.error,
              isEmpty: provider.products.isEmpty,
              emptyMessage: 'No products found',
              onRetry: provider.loadInitial,
              builder: (context) => RefreshIndicator(
                onRefresh: provider.loadInitial,
                child: GridView.builder(
                  padding: const EdgeInsets.all(16),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    mainAxisSpacing: 12,
                    crossAxisSpacing: 12,
                    childAspectRatio: 0.72,
                  ),
                  itemCount: provider.products.length,
                  itemBuilder: (context, i) => _ProductCard(product: provider.products[i]),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ProductCard extends StatelessWidget {
  final Product product;

  const _ProductCard({required this.product});

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => ProductDetailScreen(product: product)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Container(
                width: double.infinity,
                color: Theme.of(context).colorScheme.surfaceContainerHighest,
                child: product.imageUrl.isEmpty
                    ? Icon(Icons.eco, size: 40, color: Theme.of(context).colorScheme.primary)
                    : Image.network(
                        product.imageUrl,
                        fit: BoxFit.cover,
                        errorBuilder: (_, __, ___) => Icon(
                          Icons.eco,
                          size: 40,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(product.name, maxLines: 1, overflow: TextOverflow.ellipsis, style: Theme.of(context).textTheme.titleSmall),
                  const SizedBox(height: 2),
                  Text(
                    '${product.price.toStringAsFixed(2)} / ${product.unit}',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Theme.of(context).colorScheme.primary,
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

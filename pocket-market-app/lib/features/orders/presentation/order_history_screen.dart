import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../core/widgets/async_state_view.dart';
import '../domain/order.dart';
import 'order_detail_screen.dart';
import 'order_status_chip.dart';
import 'orders_provider.dart';

class OrderHistoryScreen extends StatefulWidget {
  const OrderHistoryScreen({super.key});

  @override
  State<OrderHistoryScreen> createState() => _OrderHistoryScreenState();
}

class _OrderHistoryScreenState extends State<OrderHistoryScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<OrdersProvider>().load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final provider = context.watch<OrdersProvider>();

    return Scaffold(
      appBar: AppBar(title: const Text('My Orders')),
      body: AsyncStateView(
        isLoading: provider.isLoading,
        error: provider.error,
        isEmpty: provider.orders.isEmpty,
        emptyMessage: 'No orders yet',
        onRetry: provider.load,
        builder: (context) => RefreshIndicator(
          onRefresh: provider.load,
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: provider.orders.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (context, i) => _OrderTile(order: provider.orders[i]),
          ),
        ),
      ),
    );
  }
}

class _OrderTile extends StatelessWidget {
  final Order order;

  const _OrderTile({required this.order});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => OrderDetailScreen(orderId: order.id)),
        ),
        title: Text('Order #${order.id.substring(0, 8)}'),
        subtitle: Text('${order.items.length} item(s) · ${order.createdAt.toLocal().toString().split('.').first}'),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            OrderStatusChip(status: order.status),
            const SizedBox(height: 4),
            Text(order.total.toStringAsFixed(2), style: Theme.of(context).textTheme.titleSmall),
          ],
        ),
      ),
    );
  }
}

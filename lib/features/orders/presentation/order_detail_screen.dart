import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../domain/order.dart';
import 'order_status_chip.dart';
import 'orders_provider.dart';

const _trackingSteps = [
  OrderStatus.pending,
  OrderStatus.accepted,
  OrderStatus.preparing,
  OrderStatus.outForDelivery,
  OrderStatus.delivered,
];

class OrderDetailScreen extends StatefulWidget {
  final String orderId;

  const OrderDetailScreen({super.key, required this.orderId});

  @override
  State<OrderDetailScreen> createState() => _OrderDetailScreenState();
}

class _OrderDetailScreenState extends State<OrderDetailScreen> {
  late Future<Order> _future;

  @override
  void initState() {
    super.initState();
    _future = context.read<OrdersProvider>().getDetail(widget.orderId);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Order #${widget.orderId.substring(0, 8)}')),
      body: FutureBuilder<Order>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return Center(child: Text('${snapshot.error}'));
          }
          final order = snapshot.data!;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Status', style: Theme.of(context).textTheme.titleMedium),
                  OrderStatusChip(status: order.status),
                ],
              ),
              const SizedBox(height: 16),
              if (order.status == OrderStatus.cancelled)
                Card(
                  color: Theme.of(context).colorScheme.errorContainer,
                  child: const Padding(
                    padding: EdgeInsets.all(16),
                    child: Text('This order was cancelled.'),
                  ),
                )
              else
                _TrackingStepper(current: order.status),
              const SizedBox(height: 24),
              Text('Items', style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              Card(
                child: Column(
                  children: order.items
                      .map((item) => ListTile(
                            dense: true,
                            title: Text('Qty ${item.quantity} × ${item.unitPrice.toStringAsFixed(2)}'),
                            trailing: Text(item.lineTotal.toStringAsFixed(2)),
                          ))
                      .toList(),
                ),
              ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      _totalRow(context, 'Subtotal', order.subtotal),
                      _totalRow(context, 'Delivery fee', order.deliveryFee),
                      if (order.discount > 0) _totalRow(context, 'Discount', -order.discount),
                      const Divider(),
                      _totalRow(context, 'Total', order.total, bold: true),
                      const SizedBox(height: 8),
                      _totalRow(context, 'Payment', 0, valueText: order.paymentMethod.toUpperCase()),
                      _totalRow(context, 'Payment status', 0, valueText: order.paymentStatus),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _totalRow(BuildContext context, String label, double amount, {bool bold = false, String? valueText}) {
    final style = bold
        ? Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)
        : Theme.of(context).textTheme.bodyMedium;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: style),
          Text(valueText ?? amount.toStringAsFixed(2), style: style),
        ],
      ),
    );
  }
}

class _TrackingStepper extends StatelessWidget {
  final OrderStatus current;

  const _TrackingStepper({required this.current});

  @override
  Widget build(BuildContext context) {
    final currentIndex = _trackingSteps.indexOf(current);
    final scheme = Theme.of(context).colorScheme;

    return Column(
      children: List.generate(_trackingSteps.length, (i) {
        final step = _trackingSteps[i];
        final done = i <= currentIndex;
        return Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Column(
              children: [
                Icon(
                  done ? Icons.check_circle : Icons.radio_button_unchecked,
                  color: done ? scheme.primary : scheme.outlineVariant,
                  size: 22,
                ),
                if (i != _trackingSteps.length - 1)
                  Container(
                    width: 2,
                    height: 28,
                    color: i < currentIndex ? scheme.primary : scheme.outlineVariant,
                  ),
              ],
            ),
            const SizedBox(width: 12),
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Text(
                step.label,
                style: TextStyle(
                  fontWeight: done ? FontWeight.w600 : FontWeight.normal,
                  color: done ? scheme.onSurface : scheme.outline,
                ),
              ),
            ),
          ],
        );
      }),
    );
  }
}

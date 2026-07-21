import '../../../core/api/api_client.dart';
import '../domain/order.dart';

class OrdersRepository {
  final ApiClient _client;

  OrdersRepository(this._client);

  Future<List<Order>> checkout({required String addressId, required String paymentMethod}) async {
    final json = await _client.post('/orders/checkout', body: {
      'address_id': addressId,
      'payment_method': paymentMethod,
    }) as Map<String, dynamic>;
    return (json['orders'] as List).map((e) => Order.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<List<Order>> list() async {
    final json = await _client.get('/orders');
    return (json as List).map((e) => Order.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Order> get(String id) async {
    final json = await _client.get('/orders/$id');
    return Order.fromJson(json as Map<String, dynamic>);
  }
}

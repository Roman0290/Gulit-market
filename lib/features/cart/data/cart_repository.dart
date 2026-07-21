import '../../../core/api/api_client.dart';
import '../domain/cart.dart';

class CartRepository {
  final ApiClient _client;

  CartRepository(this._client);

  Future<Cart> get() async {
    final json = await _client.get('/cart');
    return Cart.fromJson(json as Map<String, dynamic>);
  }

  Future<void> addItem(String productId, int quantity) => _client.post('/cart/items', body: {
        'product_id': productId,
        'quantity': quantity,
      });

  Future<void> updateItem(String itemId, int quantity) =>
      _client.put('/cart/items/$itemId', body: {'quantity': quantity});

  Future<void> removeItem(String itemId) => _client.delete('/cart/items/$itemId');

  Future<void> clear() => _client.delete('/cart');
}

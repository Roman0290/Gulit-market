import '../../../core/api/api_client.dart';
import '../domain/product.dart';

class ProductsRepository {
  final ApiClient _client;

  ProductsRepository(this._client);

  Future<List<Product>> list({String? categoryId, String? query}) async {
    final json = await _client.get('/products', query: {
      if (categoryId != null && categoryId.isNotEmpty) 'category': categoryId,
      if (query != null && query.isNotEmpty) 'q': query,
    });
    return (json as List).map((e) => Product.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Product> get(String id) async {
    final json = await _client.get('/products/$id');
    return Product.fromJson(json as Map<String, dynamic>);
  }

  Future<List<Category>> categories() async {
    final json = await _client.get('/categories');
    return (json as List).map((e) => Category.fromJson(e as Map<String, dynamic>)).toList();
  }
}

import '../../../core/api/api_client.dart';
import '../domain/address.dart';

class AddressesRepository {
  final ApiClient _client;

  AddressesRepository(this._client);

  Future<List<Address>> list() async {
    final json = await _client.get('/addresses');
    return (json as List).map((e) => Address.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Address> create({
    required String label,
    required String line1,
    required String city,
    bool isDefault = false,
  }) async {
    final json = await _client.post('/addresses', body: {
      'label': label,
      'line1': line1,
      'city': city,
      'is_default': isDefault,
    });
    return Address.fromJson(json as Map<String, dynamic>);
  }
}

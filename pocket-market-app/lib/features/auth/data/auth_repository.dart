import '../../../core/api/api_client.dart';
import '../../../core/api/token_store.dart';
import '../domain/user.dart';

class AuthRepository {
  final ApiClient _client;
  final TokenStore _tokenStore;

  AuthRepository({required ApiClient client, required TokenStore tokenStore})
      : _client = client,
        _tokenStore = tokenStore;

  Future<AppUser> register({
    required String name,
    required String email,
    required String phone,
    required String password,
    required String role,
  }) async {
    final json = await _client.post('/auth/register', body: {
      'name': name,
      'email': email,
      'phone': phone,
      'password': password,
      'role': role,
    });
    return AppUser.fromJson(json as Map<String, dynamic>);
  }

  Future<AppUser> login({required String email, required String password}) async {
    final json = await _client.post('/auth/login', body: {
      'email': email,
      'password': password,
    }) as Map<String, dynamic>;

    await _tokenStore.write(json['token'] as String);
    return AppUser.fromJson(json['user'] as Map<String, dynamic>);
  }

  Future<AppUser> me() async {
    final json = await _client.get('/auth/me');
    return AppUser.fromJson(json as Map<String, dynamic>);
  }

  Future<void> logout() => _tokenStore.clear();

  Future<bool> hasStoredToken() async => (await _tokenStore.read()) != null;
}

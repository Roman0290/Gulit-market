import 'package:flutter/foundation.dart';

import '../../../core/api/api_exception.dart';
import '../data/auth_repository.dart';
import '../domain/user.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthProvider extends ChangeNotifier {
  final AuthRepository _repository;

  AuthProvider(this._repository);

  AuthStatus status = AuthStatus.unknown;
  AppUser? currentUser;
  bool isBusy = false;
  String? error;

  /// Called once at app startup to see if a stored token still resolves to
  /// a valid session.
  Future<void> restoreSession() async {
    if (!await _repository.hasStoredToken()) {
      status = AuthStatus.unauthenticated;
      notifyListeners();
      return;
    }
    try {
      currentUser = await _repository.me();
      status = AuthStatus.authenticated;
    } catch (_) {
      await _repository.logout();
      status = AuthStatus.unauthenticated;
    }
    notifyListeners();
  }

  Future<bool> login(String email, String password) async {
    isBusy = true;
    error = null;
    notifyListeners();
    try {
      currentUser = await _repository.login(email: email, password: password);
      status = AuthStatus.authenticated;
      return true;
    } on ApiException catch (e) {
      error = e.message;
      return false;
    } finally {
      isBusy = false;
      notifyListeners();
    }
  }

  Future<bool> register({
    required String name,
    required String email,
    required String phone,
    required String password,
    required String role,
  }) async {
    isBusy = true;
    error = null;
    notifyListeners();
    try {
      await _repository.register(
        name: name,
        email: email,
        phone: phone,
        password: password,
        role: role,
      );
      return await login(email, password);
    } on ApiException catch (e) {
      error = e.message;
      return false;
    } finally {
      isBusy = false;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    await _repository.logout();
    currentUser = null;
    status = AuthStatus.unauthenticated;
    notifyListeners();
  }
}

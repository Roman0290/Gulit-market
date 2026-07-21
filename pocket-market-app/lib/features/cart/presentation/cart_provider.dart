import 'package:flutter/foundation.dart';

import '../../../core/api/api_exception.dart';
import '../data/cart_repository.dart';
import '../domain/cart.dart';

class CartProvider extends ChangeNotifier {
  final CartRepository _repository;

  CartProvider(this._repository);

  Cart cart = Cart.empty();
  bool isLoading = false;
  String? error;

  Future<void> load() async {
    isLoading = true;
    error = null;
    notifyListeners();
    try {
      cart = await _repository.get();
    } on ApiException catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<bool> addItem(String productId, int quantity) async {
    error = null;
    try {
      await _repository.addItem(productId, quantity);
      await load();
      return true;
    } on ApiException catch (e) {
      error = e.message;
      notifyListeners();
      return false;
    }
  }

  Future<bool> updateQuantity(String itemId, int quantity) async {
    error = null;
    try {
      await _repository.updateItem(itemId, quantity);
      await load();
      return true;
    } on ApiException catch (e) {
      error = e.message;
      notifyListeners();
      return false;
    }
  }

  Future<void> removeItem(String itemId) async {
    error = null;
    try {
      await _repository.removeItem(itemId);
      await load();
    } on ApiException catch (e) {
      error = e.message;
      notifyListeners();
    }
  }

  /// Called after a successful checkout - the server already cleared the
  /// cart, so just reset local state without another round trip.
  void reset() {
    cart = Cart.empty();
    notifyListeners();
  }
}

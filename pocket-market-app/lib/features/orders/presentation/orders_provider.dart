import 'package:flutter/foundation.dart';

import '../../../core/api/api_exception.dart';
import '../data/orders_repository.dart';
import '../domain/order.dart';

class OrdersProvider extends ChangeNotifier {
  final OrdersRepository _repository;

  OrdersProvider(this._repository);

  List<Order> orders = [];
  bool isLoading = false;
  String? error;

  Future<void> load() async {
    isLoading = true;
    error = null;
    notifyListeners();
    try {
      orders = await _repository.list();
    } on ApiException catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<Order> getDetail(String id) => _repository.get(id);

  Future<List<Order>> checkout({required String addressId, required String paymentMethod}) =>
      _repository.checkout(addressId: addressId, paymentMethod: paymentMethod);
}

import 'dart:async';

import 'package:flutter/foundation.dart' hide Category;

import '../../../core/api/api_exception.dart';
import '../data/products_repository.dart';
import '../domain/product.dart';

class ProductsProvider extends ChangeNotifier {
  final ProductsRepository _repository;

  ProductsProvider(this._repository);

  List<Product> products = [];
  List<Category> categories = [];
  String? selectedCategoryId;
  String searchQuery = '';

  bool isLoading = false;
  String? error;

  Timer? _debounce;

  Future<void> loadInitial() async {
    isLoading = true;
    error = null;
    notifyListeners();
    try {
      categories = await _repository.categories();
      products = await _repository.list();
    } on ApiException catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<void> _reload() async {
    isLoading = true;
    error = null;
    notifyListeners();
    try {
      products = await _repository.list(categoryId: selectedCategoryId, query: searchQuery);
    } on ApiException catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  void selectCategory(String? categoryId) {
    selectedCategoryId = (selectedCategoryId == categoryId) ? null : categoryId;
    _reload();
  }

  void search(String query) {
    searchQuery = query;
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), _reload);
  }

  @override
  void dispose() {
    _debounce?.cancel();
    super.dispose();
  }
}

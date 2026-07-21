class Product {
  final String id;
  final String vendorId;
  final String categoryId;
  final String name;
  final String description;
  final double price;
  final String unit;
  final int stockQuantity;
  final String imageUrl;
  final bool isActive;

  Product({
    required this.id,
    required this.vendorId,
    required this.categoryId,
    required this.name,
    required this.description,
    required this.price,
    required this.unit,
    required this.stockQuantity,
    required this.imageUrl,
    required this.isActive,
  });

  factory Product.fromJson(Map<String, dynamic> json) => Product(
        id: json['id'] as String,
        vendorId: json['vendor_id'] as String,
        categoryId: (json['category_id'] as String?) ?? '',
        name: json['name'] as String,
        description: (json['description'] as String?) ?? '',
        price: (json['price'] as num).toDouble(),
        unit: json['unit'] as String,
        stockQuantity: json['stock_quantity'] as int,
        imageUrl: (json['image_url'] as String?) ?? '',
        isActive: json['is_active'] as bool,
      );
}

class Category {
  final String id;
  final String name;
  final String iconUrl;

  Category({required this.id, required this.name, required this.iconUrl});

  factory Category.fromJson(Map<String, dynamic> json) => Category(
        id: json['id'] as String,
        name: json['name'] as String,
        iconUrl: (json['icon_url'] as String?) ?? '',
      );
}

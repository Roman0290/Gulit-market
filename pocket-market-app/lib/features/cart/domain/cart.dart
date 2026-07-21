class CartItem {
  final String id;
  final String productId;
  final String productName;
  final String vendorId;
  final double unitPrice;
  final String unit;
  final int quantity;
  final double lineTotal;

  CartItem({
    required this.id,
    required this.productId,
    required this.productName,
    required this.vendorId,
    required this.unitPrice,
    required this.unit,
    required this.quantity,
    required this.lineTotal,
  });

  factory CartItem.fromJson(Map<String, dynamic> json) => CartItem(
        id: json['id'] as String,
        productId: json['product_id'] as String,
        productName: json['product_name'] as String,
        vendorId: json['vendor_id'] as String,
        unitPrice: (json['unit_price'] as num).toDouble(),
        unit: json['unit'] as String,
        quantity: json['quantity'] as int,
        lineTotal: (json['line_total'] as num).toDouble(),
      );
}

class Cart {
  final String id;
  final List<CartItem> items;
  final double subtotal;

  Cart({required this.id, required this.items, required this.subtotal});

  factory Cart.fromJson(Map<String, dynamic> json) => Cart(
        id: (json['id'] as String?) ?? '',
        items: (json['items'] as List).map((e) => CartItem.fromJson(e as Map<String, dynamic>)).toList(),
        subtotal: (json['subtotal'] as num).toDouble(),
      );

  static Cart empty() => Cart(id: '', items: [], subtotal: 0);

  /// How many distinct vendors are represented — checkout will split into
  /// this many separate orders.
  int get vendorCount => items.map((i) => i.vendorId).toSet().length;
}

class Address {
  final String id;
  final String label;
  final String line1;
  final String city;
  final bool isDefault;

  Address({
    required this.id,
    required this.label,
    required this.line1,
    required this.city,
    required this.isDefault,
  });

  factory Address.fromJson(Map<String, dynamic> json) => Address(
        id: json['id'] as String,
        label: (json['label'] as String?) ?? '',
        line1: json['line1'] as String,
        city: json['city'] as String,
        isDefault: json['is_default'] as bool,
      );

  String get displayLabel => label.isEmpty ? line1 : label;
}

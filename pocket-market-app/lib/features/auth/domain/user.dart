class AppUser {
  final String id;
  final String name;
  final String email;
  final String phone;
  final String role;

  AppUser({
    required this.id,
    required this.name,
    required this.email,
    required this.phone,
    required this.role,
  });

  factory AppUser.fromJson(Map<String, dynamic> json) => AppUser(
        id: json['id'] as String,
        name: json['name'] as String,
        email: json['email'] as String,
        phone: json['phone'] as String,
        role: json['role'] as String,
      );

  bool get isVendor => role == 'vendor';
  bool get isCustomer => role == 'customer';
  bool get isAdmin => role == 'admin';
}

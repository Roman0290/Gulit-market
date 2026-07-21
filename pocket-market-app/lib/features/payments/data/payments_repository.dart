import '../../../core/api/api_client.dart';

class PaymentIntentResult {
  final String clientSecret;
  final String paymentIntentId;

  PaymentIntentResult({required this.clientSecret, required this.paymentIntentId});

  factory PaymentIntentResult.fromJson(Map<String, dynamic> json) => PaymentIntentResult(
        clientSecret: json['client_secret'] as String,
        paymentIntentId: json['payment_intent_id'] as String,
      );
}

class PaymentsRepository {
  final ApiClient _client;

  PaymentsRepository(this._client);

  Future<PaymentIntentResult> createIntent(String orderId) async {
    final json = await _client.post('/payments/intent', body: {'order_id': orderId});
    return PaymentIntentResult.fromJson(json as Map<String, dynamic>);
  }
}

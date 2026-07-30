package payments

import (
	"context"
	"encoding/json"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/refund"
	"github.com/stripe/stripe-go/v81/webhook"
)

const currency = "usd"

type Service struct {
	repo                *Repository
	stripeWebhookSecret string
}

func NewService(repo *Repository, stripeSecretKey, stripeWebhookSecret string) *Service {
	stripe.Key = stripeSecretKey
	return &Service{repo: repo, stripeWebhookSecret: stripeWebhookSecret}
}

type IntentResult struct {
	ClientSecret    string `json:"client_secret"`
	PaymentIntentID string `json:"payment_intent_id"`
}

func (s *Service) CreateIntent(ctx context.Context, customerID, orderID string) (*IntentResult, error) {
	order, err := s.repo.GetOrderForPayment(ctx, orderID, customerID)
	if err != nil {
		return nil, err
	}
	if order.PaymentStatus == "paid" {
		return nil, ErrOrderAlreadyPaid
	}

	amountCents := int64(order.Total*100 + 0.5)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(currency),
		Metadata: map[string]string{"order_id": orderID},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.Create(ctx, &Payment{
		OrderID:     orderID,
		Provider:    "stripe",
		ProviderRef: pi.ID,
		Amount:      order.Total,
		Status:      StatusPending,
	}); err != nil {
		return nil, err
	}

	return &IntentResult{ClientSecret: pi.ClientSecret, PaymentIntentID: pi.ID}, nil
}

// RefundOrder issues a full Stripe refund against an order's succeeded
// payment. The Stripe API call happens before the local DB update so we
// never mark something refunded that Stripe rejected.
func (s *Service) RefundOrder(ctx context.Context, orderID string) (*Payment, error) {
	payment, err := s.repo.GetSucceededPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	_, err = refund.New(&stripe.RefundParams{
		PaymentIntent: stripe.String(payment.ProviderRef),
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.MarkRefunded(ctx, payment.ID); err != nil {
		return nil, err
	}
	payment.Status = StatusRefunded
	return payment, nil
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	// IgnoreAPIVersionMismatch: the Stripe account's default API version can
	// trail the pinned stripe-go SDK version; we only read stable fields
	// (id, status) off the event so this is safe.
	event, err := webhook.ConstructEventWithOptions(payload, signatureHeader, s.stripeWebhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return err
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return err
		}
		return s.repo.MarkSucceeded(ctx, pi.ID)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return err
		}
		return s.repo.MarkFailed(ctx, pi.ID)
	}

	return nil
}

# ጉሊት Market (Gulit Market)

**ጉሊት** (*gulit*) is Amharic for "open-air market" — the neighborhood market where local farmers and vendors sell fresh produce directly to their community. **ጉሊት Market** brings that same idea online: a mobile marketplace connecting local vendors and customers for everyday essentials — vegetables, grains, oils, beans, and packaged foods — without the middlemen.

Customers get a curated, reliable shopping experience with delivery. Vendors get a digital storefront and order management without needing to be tech experts. Admins get full oversight: vendor approval, moderation, commissions, payouts, refunds, and dispute resolution.

## Features

**Customers**
- Browse, search, and filter products by category
- Server-persisted cart (survives across devices/sessions)
- Checkout with card (Stripe) or cash on delivery, coupon codes supported
- Order history and live order status tracking
- Multi-vendor carts automatically split into one order per vendor at checkout

**Vendors**
- Self-service registration, pending admin approval before going live
- Manage their own product listings and stock levels
- View and progress incoming orders through a real fulfillment lifecycle (`pending → accepted → preparing → out_for_delivery → delivered`)
- Track earnings and payout history

**Admins**
- Approve, reject, or suspend vendors
- Suspend or reactivate any user account (takes effect immediately, mid-session)
- Moderate any product listing platform-wide
- Configure commission rate, tax rate, and delivery fee
- Issue Stripe refunds on paid orders
- Record vendor payouts against their earnings ledger
- Resolve disputes raised by customers or vendors
- Manage coupons and homepage banners
- View platform analytics: revenue, order counts, top vendors, 30-day growth

## Tech Stack

| | |
|---|---|
| **Backend** | Go, [Gin](https://gin-gonic.com/), PostgreSQL, JWT auth, [Stripe](https://stripe.com/) (payments + refunds), [golang-migrate](https://github.com/golang-migrate/migrate) |
| **Frontend** | Flutter, [Provider](https://pub.dev/packages/provider) (state management), [flutter_stripe](https://pub.dev/packages/flutter_stripe) |
| **Infra** | Docker Compose (Postgres), 20 versioned SQL migrations |

The backend follows a layered architecture per feature — `handler.go` (HTTP) → `service.go` (business logic, where needed) → `repository.go` (SQL) — across 15 domain packages (auth, users, vendors, products, categories, cart, orders, payments, payouts, disputes, coupons, banners, settings, admin, addresses).

## Project Structure

```
.
├── pocket-market-api/    # Go + Gin backend
│   ├── cmd/api/           # main.go — wiring & route registration
│   ├── internal/          # one package per domain (see above)
│   ├── pkg/                # db connection, env config
│   └── migrations/        # golang-migrate SQL files
└── pocket-market-app/     # Flutter frontend (customer-facing flow)
    └── lib/
        ├── core/            # API client, theme, shared widgets
        └── features/        # auth, products, cart, orders, payments, addresses
```

## Getting Started

### Prerequisites
- Go 1.22+
- Flutter SDK
- Docker (for Postgres)
- A [Stripe](https://dashboard.stripe.com/test/apikeys) test-mode account (free)

### 1. Backend

```bash
cd pocket-market-api
cp .env.example .env   # then fill in JWT_SECRET and your Stripe test keys
docker compose up -d   # starts Postgres on :5432
migrate -path migrations -database "postgres://pocket:pocket@localhost:5432/pocket_market?sslmode=disable" up
go run ./cmd/api        # starts the API on :8080
```

To test Stripe payments/refunds locally, forward webhooks with the [Stripe CLI](https://stripe.com/docs/stripe-cli):

```bash
stripe listen --forward-to localhost:8080/api/v1/payments/webhook
```

### 2. Frontend

```bash
cd pocket-market-app
flutter pub get
flutter run -d chrome    # or a connected device/emulator
```

By default the app points at `http://localhost:8080/api/v1`. Override with `--dart-define=API_BASE_URL=...` if your API runs elsewhere (e.g. `http://10.0.2.2:8080/api/v1` for the Android emulator), and `--dart-define=STRIPE_PUBLISHABLE_KEY=pk_test_...` to enable in-app card payments.

## License

Licensed under the [Apache License 2.0](LICENSE).

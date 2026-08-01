package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/addresses"
	"github.com/romina/pocket-market-api/internal/admin"
	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/cart"
	"github.com/romina/pocket-market-api/internal/categories"
	"github.com/romina/pocket-market-api/internal/disputes"
	"github.com/romina/pocket-market-api/internal/orders"
	"github.com/romina/pocket-market-api/internal/payments"
	"github.com/romina/pocket-market-api/internal/payouts"
	"github.com/romina/pocket-market-api/internal/products"
	"github.com/romina/pocket-market-api/internal/settings"
	"github.com/romina/pocket-market-api/internal/users"
	"github.com/romina/pocket-market-api/internal/vendors"
	"github.com/romina/pocket-market-api/pkg/config"
	"github.com/romina/pocket-market-api/pkg/db"
)

func main() {
	cfg := config.Load()

	conn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	userRepo := users.NewRepository(conn)
	authService := auth.NewService(userRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService, userRepo)

	categoryRepo := categories.NewRepository(conn)
	categoryHandler := categories.NewHandler(categoryRepo)

	vendorRepo := vendors.NewRepository(conn)
	vendorService := vendors.NewService(vendorRepo)
	vendorHandler := vendors.NewHandler(vendorService)

	productRepo := products.NewRepository(conn)
	productService := products.NewService(productRepo, vendorRepo)
	productHandler := products.NewHandler(productService)

	addressRepo := addresses.NewRepository(conn)
	addressHandler := addresses.NewHandler(addressRepo)

	cartRepo := cart.NewRepository(conn)
	cartHandler := cart.NewHandler(cartRepo)

	settingsRepo := settings.NewRepository(conn)
	settingsHandler := settings.NewHandler(settingsRepo)

	payoutRepo := payouts.NewRepository(conn)
	payoutService := payouts.NewService(payoutRepo, vendorRepo)
	payoutHandler := payouts.NewHandler(payoutRepo, payoutService)

	disputeRepo := disputes.NewRepository(conn)
	disputeHandler := disputes.NewHandler(disputeRepo)

	orderRepo := orders.NewRepository(conn)
	orderService := orders.NewService(orderRepo, vendorRepo, settingsRepo)
	orderHandler := orders.NewHandler(orderService)

	paymentRepo := payments.NewRepository(conn)
	paymentService := payments.NewService(paymentRepo, cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	paymentHandler := payments.NewHandler(paymentService)

	adminRepo := admin.NewRepository(conn)
	adminHandler := admin.NewHandler(adminRepo, orderRepo, vendorRepo, userRepo, productRepo)

	router := gin.Default()

	// Permissive CORS for local development (Flutter web dev server runs on
	// a random localhost port). Tighten this to a real allowlist before
	// deploying anywhere public.
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.Static("/images", "./public/images")

	v1 := router.Group("/api/v1")
	authHandler.RegisterRoutes(v1)
	categoryHandler.RegisterRoutes(v1, authService)
	vendorHandler.RegisterRoutes(v1, authService)
	productHandler.RegisterRoutes(v1, authService)
	addressHandler.RegisterRoutes(v1, authService)
	cartHandler.RegisterRoutes(v1, authService)
	orderHandler.RegisterRoutes(v1, authService)
	paymentHandler.RegisterRoutes(v1, authService)
	adminHandler.RegisterRoutes(v1, authService)
	settingsHandler.RegisterRoutes(v1, authService)
	payoutHandler.RegisterRoutes(v1, authService)
	disputeHandler.RegisterRoutes(v1, authService)

	log.Printf("starting pocket-market-api on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

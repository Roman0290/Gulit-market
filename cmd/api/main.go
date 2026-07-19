package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/auth"
	"github.com/romina/pocket-market-api/internal/categories"
	"github.com/romina/pocket-market-api/internal/products"
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

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	authHandler.RegisterRoutes(v1)
	categoryHandler.RegisterRoutes(v1, authService)
	vendorHandler.RegisterRoutes(v1, authService)
	productHandler.RegisterRoutes(v1, authService)

	log.Printf("starting pocket-market-api on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

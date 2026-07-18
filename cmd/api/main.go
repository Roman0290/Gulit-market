package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("starting pocket-market-api on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

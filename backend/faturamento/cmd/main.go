package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/config"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/database"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "down",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "up",
		})
	})
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
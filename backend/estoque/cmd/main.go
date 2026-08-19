package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/config"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/database"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/handler"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/middleware"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/service"
)

func main() {
	cfg := config.Load("../../.env")

	db := database.Connect(cfg)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("erro ao obter conexao sql: %v", err)
	}

	produtoRepository := repository.NewProdutoRepository(db)
	produtoService := service.NewProdutoService(produtoRepository)
	produtoHandler := handler.NewProdutoHandler(produtoService)

	router := gin.New()

	router.Use(
		gin.Logger(),
		middleware.Recovery(),
	)

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

	api := router.Group("/api/v1")

	produtoHandler.RegisterRoutes(api)

	port := cfg.ServerPort

	log.Printf("[ESTOQUE] servidor rodando na porta %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("erro ao subir servidor: %v", err)
	}
}
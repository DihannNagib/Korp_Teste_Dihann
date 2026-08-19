package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/client"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/config"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/database"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/handler"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/middleware"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/service"
)

func main() {
	cfg := config.Load("../../.env")

	db := database.Connect(cfg)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf(
			"erro ao obter conexao sql: %v",
			err,
		)
	}

	estoqueClient := client.NewEstoqueClient(
		cfg.EstoqueServiceURL,
	)

	notaRepository := repository.NewNotaFiscalRepository(db)

	notaService := service.NewNotaFiscalService(
		notaRepository,
		estoqueClient,
	)

	notaHandler := handler.NewNotaFiscalHandler(
		notaService,
	)

	router := gin.New()

	router.Use(
		gin.Logger(),
		middleware.Recovery(),
	)

	router.GET("/health", func(c *gin.Context) {
		if err := sqlDB.Ping(); err != nil {
			c.JSON(
				http.StatusServiceUnavailable,
				gin.H{
					"status":   "error",
					"database": "down",
				},
			)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{
				"status":   "ok",
				"database": "up",
			},
		)
	})

	api := router.Group("/api/v1")

	notaHandler.RegisterRoutes(api)

	log.Printf(
		"[FATURAMENTO] servidor rodando na porta %s",
		cfg.ServerPort,
	)

	log.Printf(
		"[FATURAMENTO] Estoque em %s",
		cfg.EstoqueServiceURL,
	)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf(
			"erro ao subir servidor: %v",
			err,
		)
	}
}
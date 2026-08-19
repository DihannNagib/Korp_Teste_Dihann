package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("[ESTOQUE] panic recuperado: %v", recovered)

		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "erro interno inesperado no serviço de estoque",
		})
	})
}
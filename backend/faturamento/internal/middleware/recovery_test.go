package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/middleware"
)

func TestRecoveryRecuperaPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.Recovery())

	router.GET("/panic", func(c *gin.Context) {
		panic("erro inesperado")
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusInternalServerError,
		rec.Code,
	)

	assert.Contains(
		t,
		rec.Body.String(),
		"erro interno inesperado",
	)
}
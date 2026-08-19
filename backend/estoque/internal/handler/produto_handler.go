package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/dto"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/service"
)

type ProdutoHandler struct {
	service service.ProdutoService
}

func NewProdutoHandler(s service.ProdutoService) *ProdutoHandler {
	return &ProdutoHandler{
		service: s,
	}
}

func (h *ProdutoHandler) RegisterRoutes(router *gin.RouterGroup) {
	produtos := router.Group("/produtos")

	{
		produtos.POST("", h.Criar)
		produtos.GET("", h.Listar)
		produtos.GET("/:codigo", h.BuscarPorCodigo)
		produtos.PATCH("/:codigo/saldo", h.AjustarSaldo)
	}
}

func (h *ProdutoHandler) Criar(c *gin.Context) {
	var req dto.CriarProdutoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": err.Error(),
		})
		return
	}

	produto, err := h.service.CriarProduto(
		req.Codigo,
		req.Descricao,
		req.Saldo,
	)

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusCreated, produto)
}

func (h *ProdutoHandler) Listar(c *gin.Context) {
	produtos, err := h.service.ListarProdutos()

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, produtos)
}

func (h *ProdutoHandler) BuscarPorCodigo(c *gin.Context) {
	produto, err := h.service.BuscarProduto(c.Param("codigo"))

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, produto)
}

func (h *ProdutoHandler) AjustarSaldo(c *gin.Context) {
	var req dto.AjustarSaldoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": err.Error(),
		})
		return
	}

	produto, err := h.service.AjustarSaldo(
		c.Param("codigo"),
		req.Delta,
	)

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, produto)
}

func responderErro(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrProdutoNaoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"erro": err.Error(),
		})

	case errors.Is(err, repository.ErrCodigoJaExiste):
		c.JSON(http.StatusConflict, gin.H{
			"erro": err.Error(),
		})

	case errors.Is(err, domain.ErrSaldoInsuficiente):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"erro": err.Error(),
		})

	case errors.Is(err, service.ErrCodigoObrigatorio),
		errors.Is(err, service.ErrDescricaoObrigatoria),
		errors.Is(err, service.ErrSaldoInvalido),
		errors.Is(err, service.ErrDeltaInvalido):
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "erro interno no serviço de estoque",
		})
	}
}
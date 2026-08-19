package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/dto"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/service"
)

var ErrNumeroNotaInvalido = errors.New("número da nota inválido")

type NotaFiscalHandler struct {
	service service.NotaFiscalService
}

func NewNotaFiscalHandler(
	s service.NotaFiscalService,
) *NotaFiscalHandler {
	return &NotaFiscalHandler{
		service: s,
	}
}

func (h *NotaFiscalHandler) RegisterRoutes(
	router *gin.RouterGroup,
) {
	notas := router.Group("/notas")

	{
		notas.POST("", h.Criar)
		notas.GET("", h.Listar)
		notas.GET("/:numero", h.BuscarPorNumero)
		notas.POST("/:numero/imprimir", h.Imprimir)
	}
}

func (h *NotaFiscalHandler) Criar(c *gin.Context) {
	var req dto.CriarNotaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		responderErro(c, err)
		return
	}

	itens := make([]service.ItemNotaFiscalInput, len(req.Itens))

	for i, item := range req.Itens {
		itens[i] = service.ItemNotaFiscalInput{
			ProdutoCodigo: item.ProdutoCodigo,
			Quantidade:    item.Quantidade,
		}
	}

	nota, err := h.service.CriarNota(itens)
	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusCreated, nota)
}

func (h *NotaFiscalHandler) Listar(c *gin.Context) {
	notas, err := h.service.Listar()

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, notas)
}

func (h *NotaFiscalHandler) BuscarPorNumero(c *gin.Context) {
	numero, err := parseNumero(c)

	if err != nil {
		responderErro(c, err)
		return
	}

	nota, err := h.service.BuscarPorNumero(numero)

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, nota)
}

func (h *NotaFiscalHandler) Imprimir(c *gin.Context) {
	numero, err := parseNumero(c)

	if err != nil {
		responderErro(c, err)
		return
	}

	nota, err := h.service.Imprimir(numero)

	if err != nil {
		responderErro(c, err)
		return
	}

	c.JSON(http.StatusOK, nota)
}

func parseNumero(c *gin.Context) (uint, error) {
	numero, err := strconv.ParseUint(
		c.Param("numero"),
		10,
		64,
	)

	if err != nil || numero == 0 {
		return 0, ErrNumeroNotaInvalido
	}

	return uint(numero), nil
}

func responderErro(
	c *gin.Context,
	err error,
) {
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"erros": traduzirErroValidacao(ve),
			},
		)
		return
	}

	switch {
	case errors.Is(err, ErrNumeroNotaInvalido):
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"erro": err.Error(),
			},
		)

	case errors.Is(err, repository.ErrNotaNaoEncontrada):
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"erro": err.Error(),
			},
		)

	case errors.Is(err, domain.ErrNotaNaoAberta):
		c.JSON(
			http.StatusUnprocessableEntity,
			gin.H{
				"erro": err.Error(),
			},
		)

	case errors.Is(err, domain.ErrSemItens),
		errors.Is(err, domain.ErrQuantidadeInvalida),
		errors.Is(err, domain.ErrProdutoInvalido):
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"erro": err.Error(),
			},
		)

	case errors.Is(err, service.ErrSaldoInsuficienteEstoque):
		c.JSON(
			http.StatusUnprocessableEntity,
			gin.H{
				"erro": err.Error(),
			},
		)

	case errors.Is(err, service.ErrProdutoNaoEncontradoEstoque):
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"erro": err.Error(),
			},
		)

	case errors.Is(err, service.ErrFalhaComunicacaoEstoque):
		c.JSON(
			http.StatusServiceUnavailable,
			gin.H{
				"erro": err.Error(),
			},
		)

	default:
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"erro": "erro interno inesperado",
			},
		)
	}
}
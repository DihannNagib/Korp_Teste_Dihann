package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/handler"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/service"
)

type mockProdutoService struct {
	criarProdutoFn     func(string, string, int) (*domain.Produto, error)
	buscarProdutoFn    func(string) (*domain.Produto, error)
	listarProdutosFn   func() ([]domain.Produto, error)
	ajustarSaldoFn     func(string, int) (*domain.Produto, error)
}

func (m *mockProdutoService) CriarProduto(
	codigo string,
	descricao string,
	saldo int,
) (*domain.Produto, error) {
	return m.criarProdutoFn(codigo, descricao, saldo)
}

func (m *mockProdutoService) BuscarProduto(
	codigo string,
) (*domain.Produto, error) {
	return m.buscarProdutoFn(codigo)
}

func (m *mockProdutoService) ListarProdutos() ([]domain.Produto, error) {
	return m.listarProdutosFn()
}

func (m *mockProdutoService) AjustarSaldo(
	codigo string,
	delta int,
) (*domain.Produto, error) {
	return m.ajustarSaldoFn(codigo, delta)
}

func setupRouter(svc service.ProdutoService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	handler := handler.NewProdutoHandler(svc)

	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	return router
}

func TestProdutoHandler_Criar(t *testing.T) {
	svc := &mockProdutoService{
		criarProdutoFn: func(codigo, descricao string, saldo int) (*domain.Produto, error) {
			return &domain.Produto{
				ID:        1,
				Codigo:    codigo,
				Descricao: descricao,
				Saldo:     saldo,
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/produtos",
		strings.NewReader(`{
			"codigo": "PROD-001",
			"descricao": "Produto teste",
			"saldo": 10
		}`),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"codigo":"PROD-001"`)
	assert.Contains(t, rec.Body.String(), `"saldo":10`)
}

func TestProdutoHandler_CriarJSONInvalido(t *testing.T) {
	svc := &mockProdutoService{
		criarProdutoFn: func(codigo, descricao string, saldo int) (*domain.Produto, error) {
			t.Fatal("service nao deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/produtos",
		strings.NewReader(`{"codigo":`),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProdutoHandler_CriarCodigoDuplicado(t *testing.T) {
	svc := &mockProdutoService{
		criarProdutoFn: func(codigo, descricao string, saldo int) (*domain.Produto, error) {
			return nil, repository.ErrCodigoJaExiste
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/produtos",
		strings.NewReader(`{
			"codigo": "PROD-001",
			"descricao": "Produto",
			"saldo": 10
		}`),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "código de produto já cadastrado")
}

func TestProdutoHandler_Listar(t *testing.T) {
	svc := &mockProdutoService{
		listarProdutosFn: func() ([]domain.Produto, error) {
			return []domain.Produto{
				{
					ID:        1,
					Codigo:    "PROD-001",
					Descricao: "Produto 1",
					Saldo:     10,
				},
				{
					ID:        2,
					Codigo:    "PROD-002",
					Descricao: "Produto 2",
					Saldo:     20,
				},
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/produtos",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "PROD-001")
	assert.Contains(t, rec.Body.String(), "PROD-002")
}

func TestProdutoHandler_BuscarProduto(t *testing.T) {
	svc := &mockProdutoService{
		buscarProdutoFn: func(codigo string) (*domain.Produto, error) {
			assert.Equal(t, "PROD-001", codigo)

			return &domain.Produto{
				ID:        1,
				Codigo:    codigo,
				Descricao: "Produto teste",
				Saldo:     10,
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/produtos/PROD-001",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "PROD-001")
}

func TestProdutoHandler_BuscarProdutoNaoEncontrado(t *testing.T) {
	svc := &mockProdutoService{
		buscarProdutoFn: func(codigo string) (*domain.Produto, error) {
			return nil, repository.ErrProdutoNaoEncontrado
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/produtos/INEXISTENTE",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "produto não encontrado")
}

func TestProdutoHandler_AjustarSaldoReducao(t *testing.T) {
	svc := &mockProdutoService{
		ajustarSaldoFn: func(codigo string, delta int) (*domain.Produto, error) {
			assert.Equal(t, "PROD-001", codigo)
			assert.Equal(t, -3, delta)

			return &domain.Produto{
				ID:        1,
				Codigo:    codigo,
				Descricao: "Produto teste",
				Saldo:     7,
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/produtos/PROD-001/saldo",
		strings.NewReader(`{"delta":-3}`),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"saldo":7`)
}

func TestProdutoHandler_AjustarSaldoInsuficiente(t *testing.T) {
	svc := &mockProdutoService{
		ajustarSaldoFn: func(codigo string, delta int) (*domain.Produto, error) {
			return nil, domain.ErrSaldoInsuficiente
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/produtos/PROD-001/saldo",
		strings.NewReader(`{"delta":-20}`),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "saldo insuficiente")
}

func TestProdutoHandler_AjustarSaldoDeltaZero(t *testing.T) {
	svc := &mockProdutoService{
		ajustarSaldoFn: func(codigo string, delta int) (*domain.Produto, error) {
			return nil, service.ErrDeltaInvalido
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/produtos/PROD-001/saldo",
		strings.NewReader(`{"delta":0}`),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "delta não pode ser zero")
}

func TestProdutoHandler_ErroInterno(t *testing.T) {
	svc := &mockProdutoService{
		listarProdutosFn: func() ([]domain.Produto, error) {
			return nil, errors.New("erro inesperado")
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/produtos",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(
		t,
		rec.Body.String(),
		"erro interno no serviço de estoque",
	)
}
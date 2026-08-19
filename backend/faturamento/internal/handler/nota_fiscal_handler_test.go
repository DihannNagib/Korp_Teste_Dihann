package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/handler"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/service"
)

type mockNotaFiscalService struct {
	criarNotaFn       func([]service.ItemNotaFiscalInput) (*domain.NotaFiscal, error)
	buscarPorNumeroFn func(uint) (*domain.NotaFiscal, error)
	listarFn          func() ([]domain.NotaFiscal, error)
	imprimirFn        func(uint) (*domain.NotaFiscal, error)
}

func (m *mockNotaFiscalService) CriarNota(
	itens []service.ItemNotaFiscalInput,
) (*domain.NotaFiscal, error) {
	return m.criarNotaFn(itens)
}

func (m *mockNotaFiscalService) BuscarPorNumero(
	numero uint,
) (*domain.NotaFiscal, error) {
	return m.buscarPorNumeroFn(numero)
}

func (m *mockNotaFiscalService) Listar() (
	[]domain.NotaFiscal,
	error,
) {
	return m.listarFn()
}

func (m *mockNotaFiscalService) Imprimir(
	numero uint,
) (*domain.NotaFiscal, error) {
	return m.imprimirFn(numero)
}

func setupRouter(
	svc service.NotaFiscalService,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	h := handler.NewNotaFiscalHandler(svc)

	api := router.Group("/api/v1")
	h.RegisterRoutes(api)

	return router
}

func TestNotaFiscalHandler_Criar(t *testing.T) {
	svc := &mockNotaFiscalService{
		criarNotaFn: func(
			itens []service.ItemNotaFiscalInput,
		) (*domain.NotaFiscal, error) {
			require.Len(t, itens, 2)

			assert.Equal(t, "PROD-001", itens[0].ProdutoCodigo)
			assert.Equal(t, 3, itens[0].Quantidade)

			assert.Equal(t, "PROD-002", itens[1].ProdutoCodigo)
			assert.Equal(t, 5, itens[1].Quantidade)

			return &domain.NotaFiscal{
				ID:     1,
				Numero: 1,
				Status: domain.StatusAberta,
			}, nil
		},
	}

	router := setupRouter(svc)

	body := `{
		"itens": [
			{
				"produtoCodigo": "PROD-001",
				"quantidade": 3
			},
			{
				"produtoCodigo": "PROD-002",
				"quantidade": 5
			}
		]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resposta domain.NotaFiscal

	require.NoError(
		t,
		json.Unmarshal(rec.Body.Bytes(), &resposta),
	)

	assert.Equal(t, uint(1), resposta.Numero)
	assert.Equal(t, domain.StatusAberta, resposta.Status)
}

func TestNotaFiscalHandler_CriarPayloadInvalido(t *testing.T) {
	svc := &mockNotaFiscalService{
		criarNotaFn: func(
			[]service.ItemNotaFiscalInput,
		) (*domain.NotaFiscal, error) {
			t.Fatal("service não deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	body := `{
		"itens": []
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotaFiscalHandler_CriarQuantidadeInvalida(t *testing.T) {
	svc := &mockNotaFiscalService{
		criarNotaFn: func(
			[]service.ItemNotaFiscalInput,
		) (*domain.NotaFiscal, error) {
			t.Fatal("service não deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	body := `{
		"itens": [
			{
				"produtoCodigo": "PROD-001",
				"quantidade": 0
			}
		]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotaFiscalHandler_Listar(t *testing.T) {
	svc := &mockNotaFiscalService{
		listarFn: func() ([]domain.NotaFiscal, error) {
			return []domain.NotaFiscal{
				{
					ID:     1,
					Numero: 1,
					Status: domain.StatusAberta,
				},
				{
					ID:     2,
					Numero: 2,
					Status: domain.StatusFechada,
				},
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/notas",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resposta []domain.NotaFiscal

	require.NoError(
		t,
		json.Unmarshal(rec.Body.Bytes(), &resposta),
	)

	assert.Len(t, resposta, 2)
}

func TestNotaFiscalHandler_BuscarPorNumero(t *testing.T) {
	svc := &mockNotaFiscalService{
		buscarPorNumeroFn: func(
			numero uint,
		) (*domain.NotaFiscal, error) {
			assert.Equal(t, uint(10), numero)

			return &domain.NotaFiscal{
				ID:     10,
				Numero: 10,
				Status: domain.StatusAberta,
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/notas/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNotaFiscalHandler_BuscarPorNumeroNaoEncontrada(t *testing.T) {
	svc := &mockNotaFiscalService{
		buscarPorNumeroFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			return nil, repository.ErrNotaNaoEncontrada
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/notas/999",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNotaFiscalHandler_BuscarNumeroInvalido(t *testing.T) {
	svc := &mockNotaFiscalService{
		buscarPorNumeroFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			t.Fatal("service não deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/notas/abc",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotaFiscalHandler_BuscarNumeroZero(t *testing.T) {
	svc := &mockNotaFiscalService{
		buscarPorNumeroFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			t.Fatal("service não deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/notas/0",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotaFiscalHandler_Imprimir(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			numero uint,
		) (*domain.NotaFiscal, error) {
			assert.Equal(t, uint(5), numero)

			return &domain.NotaFiscal{
				ID:     5,
				Numero: 5,
				Status: domain.StatusFechada,
			}, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/5/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNotaFiscalHandler_ImprimirNotaNaoAberta(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			return nil, domain.ErrNotaNaoAberta
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/1/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusUnprocessableEntity,
		rec.Code,
	)
}

func TestNotaFiscalHandler_ImprimirSaldoInsuficienteEstoque(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			return nil, service.ErrSaldoInsuficienteEstoque
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/1/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusUnprocessableEntity,
		rec.Code,
	)
}

func TestNotaFiscalHandler_ImprimirProdutoNaoEncontradoEstoque(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			return nil, service.ErrProdutoNaoEncontradoEstoque
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/1/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusNotFound,
		rec.Code,
	)
}

func TestNotaFiscalHandler_ImprimirFalhaComunicacaoEstoque(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			return nil, service.ErrFalhaComunicacaoEstoque
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/1/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusServiceUnavailable,
		rec.Code,
	)
}

func TestNotaFiscalHandler_ImprimirNumeroInvalido(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			t.Fatal("service não deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/abc/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		rec.Code,
	)
}

func TestNotaFiscalHandler_ImprimirNumeroZero(t *testing.T) {
	svc := &mockNotaFiscalService{
		imprimirFn: func(
			uint,
		) (*domain.NotaFiscal, error) {
			t.Fatal("service não deveria ser chamado")
			return nil, nil
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/notas/0/imprimir",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		rec.Code,
	)
}

func TestNotaFiscalHandler_PropagaErroInternoComo500(t *testing.T) {
	svc := &mockNotaFiscalService{
		listarFn: func() ([]domain.NotaFiscal, error) {
			return nil, errors.New("erro inesperado")
		},
	}

	router := setupRouter(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/notas",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/client"
)

func TestEstoqueClient_BaixarItem(t *testing.T) {
	var requisicoes int

	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requisicoes++

			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(
				t,
				"/api/v1/produtos/PROD-001/saldo",
				r.URL.Path,
			)

			w.WriteHeader(http.StatusOK)
		}),
	)

	defer server.Close()

	estoqueClient := client.NewEstoqueClient(server.URL)

	err := estoqueClient.BaixarItem(
		"PROD-001",
		3,
	)

	require.NoError(t, err)
	assert.Equal(t, 1, requisicoes)
}

func TestEstoqueClient_BaixarItemSaldoInsuficiente(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}),
	)

	defer server.Close()

	estoqueClient := client.NewEstoqueClient(server.URL)

	err := estoqueClient.BaixarItem(
		"PROD-001",
		10,
	)

	assert.ErrorIs(
		t,
		err,
		client.ErrSaldoInsuficiente,
	)
}

func TestEstoqueClient_BaixarItemProdutoNaoEncontrado(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	defer server.Close()

	estoqueClient := client.NewEstoqueClient(server.URL)

	err := estoqueClient.BaixarItem(
		"PROD-999",
		1,
	)

	assert.ErrorIs(
		t,
		err,
		client.ErrProdutoNaoEncontrado,
	)
}

func TestEstoqueClient_BaixarItemEstoqueIndisponivel(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			http.Error(
				w,
				"erro interno",
				http.StatusInternalServerError,
			)
		}),
	)

	defer server.Close()

	estoqueClient := client.NewEstoqueClient(server.URL)

	err := estoqueClient.BaixarItem(
		"PROD-001",
		1,
	)

	assert.ErrorIs(
		t,
		err,
		client.ErrEstoqueIndisponivel,
	)
}

func TestEstoqueClient_BaixarItemServidorIndisponivel(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
		},
		),
	)

	serverURL := server.URL
	server.Close()

	estoqueClient := client.NewEstoqueClient(serverURL)

	err := estoqueClient.BaixarItem(
		"PROD-001",
		1,
	)

	assert.ErrorIs(
		t,
		err,
		client.ErrEstoqueIndisponivel,
	)
}

func TestEstoqueClient_BaixarItemEnviaDeltaNegativo(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			assert.Equal(
				t,
				"/api/v1/produtos/PROD-001/saldo",
				r.URL.Path,
			)

			assert.Equal(
				t,
				http.MethodPatch,
				r.Method,
			)

			w.WriteHeader(http.StatusOK)
		}),
	)

	defer server.Close()

	estoqueClient := client.NewEstoqueClient(server.URL)

	err := estoqueClient.BaixarItem(
		"PROD-001",
		3,
	)

	require.NoError(t, err)
}
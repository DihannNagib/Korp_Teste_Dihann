package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
)

func TestProduto_AjustarSaldoReducao(t *testing.T) {
	produto := domain.Produto{
		Codigo: "PROD-001",
		Saldo: 10,
	}

	err := produto.AjustarSaldo(-3)

	require.NoError(t, err)
	assert.Equal(t, 7, produto.Saldo)
}

func TestProduto_AjustarSaldoIncremento(t *testing.T) {
	produto := domain.Produto{
		Codigo: "PROD-001",
		Saldo: 10,
	}

	err := produto.AjustarSaldo(5)

	require.NoError(t, err)
	assert.Equal(t, 15, produto.Saldo)
}

func TestProduto_AjustarSaldoInsuficiente(t *testing.T) {
	produto := domain.Produto{
		Codigo: "PROD-001",
		Saldo: 5,
	}

	err := produto.AjustarSaldo(-6)

	require.ErrorIs(t, err, domain.ErrSaldoInsuficiente)
	assert.Equal(t, 5, produto.Saldo)
}
package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
)

func TestNovaNotaFiscal_IniciaAberta(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	assert.Equal(t, domain.StatusAberta, nota.Status)
	assert.Empty(t, nota.Itens)
}

func TestAdicionarItem_Sucesso(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	err := nota.AdicionarItem("PROD-001", 2)

	require.NoError(t, err)
	require.Len(t, nota.Itens, 1)

	assert.Equal(t, "PROD-001", nota.Itens[0].ProdutoCodigo)
	assert.Equal(t, 2, nota.Itens[0].Quantidade)
}

func TestAdicionarItem_CodigoProdutoInvalido(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	err := nota.AdicionarItem("", 2)

	assert.ErrorIs(t, err, domain.ErrProdutoInvalido)
}

func TestAdicionarItem_QuantidadeZero(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	err := nota.AdicionarItem("PROD-001", 0)

	assert.ErrorIs(t, err, domain.ErrQuantidadeInvalida)
}

func TestAdicionarItem_QuantidadeNegativa(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	err := nota.AdicionarItem("PROD-001", -1)

	assert.ErrorIs(t, err, domain.ErrQuantidadeInvalida)
}

func TestAdicionarItem_NotaFechada(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	require.NoError(t, nota.Fechar())

	err := nota.AdicionarItem("PROD-001", 1)

	assert.ErrorIs(t, err, domain.ErrNotaNaoAberta)
}

func TestPodeSerImpressa_SemItens(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	err := nota.PodeSerImpressa()

	assert.ErrorIs(t, err, domain.ErrSemItens)
}

func TestPodeSerImpressa_NotaFechada(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	require.NoError(t, nota.AdicionarItem("PROD-001", 1))
	require.NoError(t, nota.Fechar())

	err := nota.PodeSerImpressa()

	assert.ErrorIs(t, err, domain.ErrNotaNaoAberta)
}

func TestPodeSerImpressa_Sucesso(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	require.NoError(t, nota.AdicionarItem("PROD-001", 1))

	err := nota.PodeSerImpressa()

	assert.NoError(t, err)
}

func TestFechar_Sucesso(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	err := nota.Fechar()

	require.NoError(t, err)
	assert.Equal(t, domain.StatusFechada, nota.Status)
}

func TestFechar_NotaJaFechada(t *testing.T) {
	nota := domain.NovaNotaFiscal()

	require.NoError(t, nota.Fechar())

	err := nota.Fechar()

	assert.ErrorIs(t, err, domain.ErrNotaNaoAberta)
}
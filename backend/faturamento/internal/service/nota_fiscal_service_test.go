package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/service"
)

type mockNotaFiscalRepository struct {
	notas map[uint]*domain.NotaFiscal

	createCalls          int
	findByNumeroCalls    int
	findAllCalls         int
	atualizarStatusCalls int

	createError          error
	findByNumeroError    error
	findAllError         error
	atualizarStatusError error
}

func newMockNotaFiscalRepository() *mockNotaFiscalRepository {
	return &mockNotaFiscalRepository{
		notas: make(map[uint]*domain.NotaFiscal),
	}
}

func (m *mockNotaFiscalRepository) Create(
	nota *domain.NotaFiscal,
) error {
	m.createCalls++

	if m.createError != nil {
		return m.createError
	}

	id := uint(len(m.notas) + 1)

	nota.ID = id
	nota.Numero = id

	m.notas[nota.Numero] = nota

	return nil
}

func (m *mockNotaFiscalRepository) FindByNumero(
	numero uint,
) (*domain.NotaFiscal, error) {
	m.findByNumeroCalls++

	if m.findByNumeroError != nil {
		return nil, m.findByNumeroError
	}

	nota, ok := m.notas[numero]
	if !ok {
		return nil, repository.ErrNotaNaoEncontrada
	}

	return nota, nil
}

func (m *mockNotaFiscalRepository) FindAll() ([]domain.NotaFiscal, error) {
	m.findAllCalls++

	if m.findAllError != nil {
		return nil, m.findAllError
	}

	notas := make([]domain.NotaFiscal, 0, len(m.notas))

	for _, nota := range m.notas {
		notas = append(notas, *nota)
	}

	return notas, nil
}

func (m *mockNotaFiscalRepository) AtualizarStatus(
	nota *domain.NotaFiscal,
) error {
	m.atualizarStatusCalls++

	if m.atualizarStatusError != nil {
		return m.atualizarStatusError
	}

	m.notas[nota.Numero] = nota

	return nil
}

type mockEstoqueGateway struct {
	itensRecebidos []domain.ItemNotaFiscal
	erro           error
}

func newMockEstoqueGateway() *mockEstoqueGateway {
	return &mockEstoqueGateway{
		itensRecebidos: make([]domain.ItemNotaFiscal, 0),
	}
}

func (m *mockEstoqueGateway) BaixarItens(
	itens []domain.ItemNotaFiscal,
) error {
	m.itensRecebidos = append(
		m.itensRecebidos,
		itens...,
	)

	return m.erro
}

func criarService() (
	service.NotaFiscalService,
	*mockNotaFiscalRepository,
	*mockEstoqueGateway,
) {
	repo := newMockNotaFiscalRepository()
	estoque := newMockEstoqueGateway()

	svc := service.NewNotaFiscalService(
		repo,
		estoque,
	)

	return svc, repo, estoque
}

func TestNotaFiscalService_CriarNota(t *testing.T) {
	svc, repo, _ := criarService()

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "PROD-001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "PROD-002",
			Quantidade:    5,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, nota)

	assert.Equal(t, domain.StatusAberta, nota.Status)
	assert.Equal(t, uint(1), nota.Numero)
	assert.Len(t, nota.Itens, 2)
	assert.Equal(t, 1, repo.createCalls)

	assert.Equal(t, "PROD-001", nota.Itens[0].ProdutoCodigo)
	assert.Equal(t, 3, nota.Itens[0].Quantidade)

	assert.Equal(t, "PROD-002", nota.Itens[1].ProdutoCodigo)
	assert.Equal(t, 5, nota.Itens[1].Quantidade)
}

func TestNotaFiscalService_CriarNotaValidaItens(t *testing.T) {
	tests := []struct {
		name     string
		item     service.ItemNotaFiscalInput
		expected error
	}{
		{
			name: "produto inválido",
			item: service.ItemNotaFiscalInput{
				ProdutoCodigo: "   ",
				Quantidade:    1,
			},
			expected: domain.ErrProdutoInvalido,
		},
		{
			name: "quantidade inválida",
			item: service.ItemNotaFiscalInput{
				ProdutoCodigo: "PROD-001",
				Quantidade:    0,
			},
			expected: domain.ErrQuantidadeInvalida,
		},
		{
			name: "quantidade negativa",
			item: service.ItemNotaFiscalInput{
				ProdutoCodigo: "PROD-001",
				Quantidade:    -1,
			},
			expected: domain.ErrQuantidadeInvalida,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _ := criarService()

			nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
				tt.item,
			})

			assert.Nil(t, nota)
			assert.ErrorIs(t, err, tt.expected)
			assert.Equal(t, 0, repo.createCalls)
		})
	}
}

func TestNotaFiscalService_CriarNotaPropagaErroDoRepository(t *testing.T) {
	svc, repo, _ := criarService()

	erroBanco := errors.New("erro de banco")
	repo.createError = erroBanco

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "PROD-001",
			Quantidade:    1,
		},
	})

	assert.Nil(t, nota)
	assert.ErrorIs(t, err, erroBanco)
	assert.Equal(t, 1, repo.createCalls)
}

func TestNotaFiscalService_BuscarPorNumero(t *testing.T) {
	svc, repo, _ := criarService()

	nota := domain.NovaNotaFiscal()
	require.NoError(t, repo.Create(nota))

	encontrada, err := svc.BuscarPorNumero(nota.Numero)

	require.NoError(t, err)
	require.NotNil(t, encontrada)

	assert.Equal(t, nota.Numero, encontrada.Numero)
	assert.Equal(t, 1, repo.findByNumeroCalls)
}

func TestNotaFiscalService_BuscarPorNumeroPropagaErro(t *testing.T) {
	svc, repo, _ := criarService()

	erro := repository.ErrNotaNaoEncontrada
	repo.findByNumeroError = erro

	nota, err := svc.BuscarPorNumero(999)

	assert.Nil(t, nota)
	assert.ErrorIs(t, err, erro)
	assert.Equal(t, 1, repo.findByNumeroCalls)
}

func TestNotaFiscalService_Listar(t *testing.T) {
	svc, repo, _ := criarService()

	nota1 := domain.NovaNotaFiscal()
	nota2 := domain.NovaNotaFiscal()

	require.NoError(t, repo.Create(nota1))
	require.NoError(t, repo.Create(nota2))

	notas, err := svc.Listar()

	require.NoError(t, err)
	require.Len(t, notas, 2)

	assert.Equal(t, 1, repo.findAllCalls)
}

func TestNotaFiscalService_ListarPropagaErro(t *testing.T) {
	svc, repo, _ := criarService()

	erro := errors.New("erro ao consultar notas")
	repo.findAllError = erro

	notas, err := svc.Listar()

	assert.Nil(t, notas)
	assert.ErrorIs(t, err, erro)
	assert.Equal(t, 1, repo.findAllCalls)
}

func TestNotaFiscalService_Imprimir(t *testing.T) {
	svc, repo, estoque := criarService()

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "PROD-001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "PROD-002",
			Quantidade:    5,
		},
	})

	require.NoError(t, err)

	impressa, err := svc.Imprimir(nota.Numero)

	require.NoError(t, err)
	require.NotNil(t, impressa)

	assert.Equal(t, domain.StatusFechada, impressa.Status)
	assert.Equal(t, 1, repo.atualizarStatusCalls)

	require.Len(t, estoque.itensRecebidos, 2)

	assert.Equal(t, "PROD-001", estoque.itensRecebidos[0].ProdutoCodigo)
	assert.Equal(t, 3, estoque.itensRecebidos[0].Quantidade)

	assert.Equal(t, "PROD-002", estoque.itensRecebidos[1].ProdutoCodigo)
	assert.Equal(t, 5, estoque.itensRecebidos[1].Quantidade)
}

func TestNotaFiscalService_ImprimirNotaSemItens(t *testing.T) {
	svc, repo, estoque := criarService()

	nota := domain.NovaNotaFiscal()
	require.NoError(t, repo.Create(nota))

	resultado, err := svc.Imprimir(nota.Numero)

	assert.Nil(t, resultado)
	assert.ErrorIs(t, err, domain.ErrSemItens)

	assert.Empty(t, estoque.itensRecebidos)
	assert.Equal(t, 0, repo.atualizarStatusCalls)
}

func TestNotaFiscalService_ImprimirNotaFechada(t *testing.T) {
	svc, repo, estoque := criarService()

	nota := domain.NovaNotaFiscal()

	require.NoError(
		t,
		nota.AdicionarItem("PROD-001", 2),
	)

	require.NoError(t, repo.Create(nota))
	require.NoError(t, nota.Fechar())
	require.NoError(t, repo.AtualizarStatus(nota))

	resultado, err := svc.Imprimir(nota.Numero)

	assert.Nil(t, resultado)
	assert.ErrorIs(t, err, domain.ErrNotaNaoAberta)

	assert.Empty(t, estoque.itensRecebidos)
	assert.Equal(t, 1, repo.atualizarStatusCalls)
}

func TestNotaFiscalService_ImprimirEstoqueFalha(t *testing.T) {
	svc, repo, estoque := criarService()

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "PROD-001",
			Quantidade:    3,
		},
	})

	require.NoError(t, err)

	erroEstoque := errors.New("saldo insuficiente")
	estoque.erro = erroEstoque

	resultado, err := svc.Imprimir(nota.Numero)

	assert.Nil(t, resultado)

	assert.ErrorIs(t, err, service.ErrBaixaEstoque)
	assert.ErrorIs(t, err, erroEstoque)

	require.Len(t, estoque.itensRecebidos, 1)

	assert.Equal(
		t,
		"PROD-001",
		estoque.itensRecebidos[0].ProdutoCodigo,
	)

	assert.Equal(
		t,
		3,
		estoque.itensRecebidos[0].Quantidade,
	)

	assert.Equal(t, 0, repo.atualizarStatusCalls)
	assert.Equal(t, domain.StatusAberta, nota.Status)
}

func TestNotaFiscalService_ImprimirNaoFechaQuandoEstoqueFalha(t *testing.T) {
	svc, repo, estoque := criarService()

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "PROD-001",
			Quantidade:    3,
		},
	})

	require.NoError(t, err)

	estoque.erro = errors.New("saldo insuficiente")

	resultado, err := svc.Imprimir(nota.Numero)

	assert.Nil(t, resultado)
	assert.ErrorIs(t, err, service.ErrBaixaEstoque)

	assert.Equal(t, domain.StatusAberta, nota.Status)
	assert.Equal(t, 0, repo.atualizarStatusCalls)
}

func TestNotaFiscalService_ImprimirNaoFechaSeAtualizacaoFalhar(
	t *testing.T,
) {
	svc, repo, estoque := criarService()

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "PROD-001",
			Quantidade:    2,
		},
	})

	require.NoError(t, err)

	erroAtualizacao := errors.New("erro ao atualizar nota")
	repo.atualizarStatusError = erroAtualizacao

	resultado, err := svc.Imprimir(nota.Numero)

	assert.Nil(t, resultado)
	assert.ErrorIs(t, err, erroAtualizacao)

	assert.Len(t, estoque.itensRecebidos, 1)

	// O domínio já mudou para FECHADA antes da persistência.
	assert.Equal(t, domain.StatusFechada, nota.Status)

	assert.Equal(t, 1, repo.atualizarStatusCalls)
}

func TestNotaFiscalService_ImprimirPropagaErroDaBusca(
	t *testing.T,
) {
	svc, repo, estoque := criarService()

	erro := repository.ErrNotaNaoEncontrada
	repo.findByNumeroError = erro

	resultado, err := svc.Imprimir(999)

	assert.Nil(t, resultado)
	assert.ErrorIs(t, err, erro)

	assert.Empty(t, estoque.itensRecebidos)
	assert.Equal(t, 0, repo.atualizarStatusCalls)
}

func estoqueCalls(estoque *mockEstoqueGateway) int {
	if len(estoque.itensRecebidos) == 0 {
		return 0
	}

	return 1
}

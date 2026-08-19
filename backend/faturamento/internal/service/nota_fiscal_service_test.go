package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/client"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/service"
)

type mockRepo struct {
	notas      map[uint]*domain.NotaFiscal
	proxNumero uint

	createErr         error
	findByNumeroErr   error
	findAllErr        error
	atualizarStatusErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		notas:      make(map[uint]*domain.NotaFiscal),
		proxNumero: 1,
	}
}

func (m *mockRepo) Create(n *domain.NotaFiscal) error {
	if m.createErr != nil {
		return m.createErr
	}

	n.ID = m.proxNumero
	n.Numero = m.proxNumero

	m.notas[n.Numero] = n
	m.proxNumero++

	return nil
}

func (m *mockRepo) FindByNumero(numero uint) (*domain.NotaFiscal, error) {
	if m.findByNumeroErr != nil {
		return nil, m.findByNumeroErr
	}

	nota, ok := m.notas[numero]

	if !ok {
		return nil, repository.ErrNotaNaoEncontrada
	}

	return nota, nil
}

func (m *mockRepo) FindAll() ([]domain.NotaFiscal, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}

	var notas []domain.NotaFiscal

	for _, nota := range m.notas {
		notas = append(notas, *nota)
	}

	return notas, nil
}

func (m *mockRepo) AtualizarStatus(n *domain.NotaFiscal) error {
	if m.atualizarStatusErr != nil {
		return m.atualizarStatusErr
	}

	m.notas[n.Numero].Status = n.Status

	return nil
}

type chamada struct {
	tipo       string
	codigo     string
	quantidade int
}

type mockEstoque struct {
	falharNoItem int
	erro         error

	baixas   int
	chamadas []chamada

	estornoErr error
}

func (m *mockEstoque) BaixarItem(
	codigo string,
	quantidade int,
) error {
	m.chamadas = append(
		m.chamadas,
		chamada{
			tipo:       "baixa",
			codigo:     codigo,
			quantidade: quantidade,
		},
	)

	if m.baixas == m.falharNoItem {
		m.baixas++
		return m.erro
	}

	m.baixas++

	return nil
}

func (m *mockEstoque) EstornarItem(
	codigo string,
	quantidade int,
) error {
	m.chamadas = append(
		m.chamadas,
		chamada{
			tipo:       "estorno",
			codigo:     codigo,
			quantidade: quantidade,
		},
	)

	return m.estornoErr
}

// ---------------------------------------------------------
// CriarNota
// ---------------------------------------------------------

func TestCriarNota_Sucesso(t *testing.T) {
	repo := newMockRepo()
	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "P002",
			Quantidade:    5,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, nota)

	assert.Equal(t, uint(1), nota.ID)
	assert.Equal(t, uint(1), nota.Numero)
	assert.Equal(t, domain.StatusAberta, nota.Status)

	require.Len(t, nota.Itens, 2)

	assert.Equal(t, "P001", nota.Itens[0].ProdutoCodigo)
	assert.Equal(t, 3, nota.Itens[0].Quantidade)

	assert.Equal(t, "P002", nota.Itens[1].ProdutoCodigo)
	assert.Equal(t, 5, nota.Itens[1].Quantidade)
}

func TestCriarNota_ItemInvalidoNaoPersiste(t *testing.T) {
	repo := newMockRepo()
	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	_, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "P002",
			Quantidade:    0,
		},
	})

	require.Error(t, err)

	assert.Len(t, repo.notas, 0)
}

// ---------------------------------------------------------
// BuscarPorNumero
// ---------------------------------------------------------

func TestBuscarPorNumero(t *testing.T) {
	repo := newMockRepo()
	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	criada, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    2,
		},
	})

	require.NoError(t, err)

	encontrada, err := svc.BuscarPorNumero(criada.Numero)

	require.NoError(t, err)
	require.NotNil(t, encontrada)

	assert.Equal(t, criada.Numero, encontrada.Numero)
	assert.Equal(t, domain.StatusAberta, encontrada.Status)
}

func TestBuscarPorNumeroNaoEncontrada(t *testing.T) {
	repo := newMockRepo()
	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	_, err := svc.BuscarPorNumero(999)

	assert.ErrorIs(
		t,
		err,
		repository.ErrNotaNaoEncontrada,
	)
}

// ---------------------------------------------------------
// Listar
// ---------------------------------------------------------

func TestListar(t *testing.T) {
	repo := newMockRepo()
	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	_, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    1,
		},
	})

	require.NoError(t, err)

	_, err = svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P002",
			Quantidade:    2,
		},
	})

	require.NoError(t, err)

	notas, err := svc.Listar()

	require.NoError(t, err)
	assert.Len(t, notas, 2)
}

// ---------------------------------------------------------
// Imprimir
// ---------------------------------------------------------

func TestImprimir_Sucesso(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "P002",
			Quantidade:    5,
		},
	})

	require.NoError(t, err)

	impressa, err := svc.Imprimir(nota.Numero)

	require.NoError(t, err)
	require.NotNil(t, impressa)

	assert.Equal(
		t,
		domain.StatusFechada,
		impressa.Status,
	)

	require.Len(t, estoque.chamadas, 2)

	assert.Equal(
		t,
		chamada{"baixa", "P001", 3},
		estoque.chamadas[0],
	)

	assert.Equal(
		t,
		chamada{"baixa", "P002", 5},
		estoque.chamadas[1],
	)
}

func TestImprimir_FalhaProdutoNaoEncontradoCompensa(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: 1,
		erro:         client.ErrProdutoNaoEncontrado,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "P002",
			Quantidade:    1,
		},
	})

	require.NoError(t, err)

	_, err = svc.Imprimir(nota.Numero)

	require.Error(t, err)

	assert.ErrorIs(
		t,
		err,
		service.ErrProdutoNaoEncontradoEstoque,
	)

	notaAtual, err := repo.FindByNumero(nota.Numero)

	require.NoError(t, err)

	assert.Equal(
		t,
		domain.StatusAberta,
		notaAtual.Status,
	)

	// Baixa P001
	// Baixa P002 -> falha
	// Estorno P001
	require.Len(t, estoque.chamadas, 3)

	assert.Equal(
		t,
		chamada{"baixa", "P001", 3},
		estoque.chamadas[0],
	)

	assert.Equal(
		t,
		chamada{"baixa", "P002", 1},
		estoque.chamadas[1],
	)

	assert.Equal(
		t,
		chamada{"estorno", "P001", 3},
		estoque.chamadas[2],
	)
}

func TestImprimir_SaldoInsuficiente(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: 0,
		erro:         client.ErrSaldoInsuficiente,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    10,
		},
	})

	require.NoError(t, err)

	_, err = svc.Imprimir(nota.Numero)

	require.Error(t, err)

	assert.ErrorIs(
		t,
		err,
		service.ErrSaldoInsuficienteEstoque,
	)

	notaAtual, err := repo.FindByNumero(nota.Numero)

	require.NoError(t, err)

	assert.Equal(
		t,
		domain.StatusAberta,
		notaAtual.Status,
	)

	// Como a primeira baixa falhou,
	// nenhum estorno deve ser realizado.
	require.Len(t, estoque.chamadas, 1)

	assert.Equal(
		t,
		chamada{"baixa", "P001", 10},
		estoque.chamadas[0],
	)
}

func TestImprimir_ErroComunicacaoEstoque(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: 0,
		erro:         client.ErrEstoqueIndisponivel,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    2,
		},
	})

	require.NoError(t, err)

	_, err = svc.Imprimir(nota.Numero)

	require.Error(t, err)

	assert.ErrorIs(
		t,
		err,
		service.ErrFalhaComunicacaoEstoque,
	)

	notaAtual, err := repo.FindByNumero(nota.Numero)

	require.NoError(t, err)

	assert.Equal(
		t,
		domain.StatusAberta,
		notaAtual.Status,
	)
}

func TestImprimir_NotaJaFechada(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    2,
		},
	})

	require.NoError(t, err)

	// Primeira impressão.
	_, err = svc.Imprimir(nota.Numero)

	require.NoError(t, err)

	estoque.chamadas = nil

	// Segunda impressão.
	_, err = svc.Imprimir(nota.Numero)

	require.ErrorIs(
		t,
		err,
		domain.ErrNotaNaoAberta,
	)

	// Não deve chamar o estoque novamente.
	assert.Empty(t, estoque.chamadas)
}

func TestImprimir_NotaNaoEncontrada(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	_, err := svc.Imprimir(999)

	require.Error(t, err)

	assert.ErrorIs(
		t,
		err,
		repository.ErrNotaNaoEncontrada,
	)

	assert.Empty(t, estoque.chamadas)
}

func TestImprimir_RetryAposCompensacao(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: 1,
		erro:         client.ErrProdutoNaoEncontrado,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    3,
		},
		{
			ProdutoCodigo: "P002",
			Quantidade:    1,
		},
	})

	require.NoError(t, err)

	// Primeira tentativa falha.
	_, err = svc.Imprimir(nota.Numero)

	require.Error(t, err)

	// Segunda tentativa deve funcionar.
	estoque.falharNoItem = -1
	estoque.baixas = 0
	estoque.chamadas = nil

	impressa, err := svc.Imprimir(nota.Numero)

	require.NoError(t, err)

	assert.Equal(
		t,
		domain.StatusFechada,
		impressa.Status,
	)

	require.Len(t, estoque.chamadas, 2)

	assert.Equal(
		t,
		chamada{"baixa", "P001", 3},
		estoque.chamadas[0],
	)

	assert.Equal(
		t,
		chamada{"baixa", "P002", 1},
		estoque.chamadas[1],
	)
}

// ---------------------------------------------------------
// Falhas de persistência
// ---------------------------------------------------------

func TestCriarNota_ErroRepository(t *testing.T) {
	repo := newMockRepo()

	repo.createErr = errors.New("erro banco")

	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	_, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade:    1,
		},
	})

	require.Error(t, err)

	assert.Contains(
		t,
		err.Error(),
		"erro ao criar nota fiscal",
	)
}

func TestImprimir_ErroAtualizarStatusCompensa(t *testing.T) {
	repo := newMockRepo()

	estoque := &mockEstoque{
		falharNoItem: -1,
	}

	svc := service.NewNotaFiscalService(repo, estoque)

	nota, err := svc.CriarNota([]service.ItemNotaFiscalInput{
		{
			ProdutoCodigo: "P001",
			Quantidade: 3,
		},
	})

	require.NoError(t, err)

	repo.atualizarStatusErr = errors.New("erro ao atualizar status")

	_, err = svc.Imprimir(nota.Numero)

	require.Error(t, err)

	assert.Contains(
		t,
		err.Error(),
		"erro ao atualizar status",
	)

	require.Len(t, estoque.chamadas, 2)

	assert.Equal(
		t,
		chamada{"baixa", "P001", 3},
		estoque.chamadas[0],
	)

	assert.Equal(
		t,
		chamada{"estorno", "P001", 3},
		estoque.chamadas[1],
	)
}
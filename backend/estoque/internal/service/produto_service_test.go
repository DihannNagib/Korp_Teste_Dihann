package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/repository"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/service"
)

type mockProdutoRepository struct {
	produtos map[string]*domain.Produto
}

func newMockProdutoRepository() *mockProdutoRepository {
	return &mockProdutoRepository{
		produtos: make(map[string]*domain.Produto),
	}
}

func (m *mockProdutoRepository) Create(
	produto *domain.Produto,
) error {

	if _, exists := m.produtos[produto.Codigo]; exists {
		return repository.ErrCodigoJaExiste
	}

	m.produtos[produto.Codigo] = produto

	return nil
}

func (m *mockProdutoRepository) FindByCodigo(
	codigo string,
) (*domain.Produto, error) {

	produto, exists := m.produtos[codigo]

	if !exists {
		return nil, repository.ErrProdutoNaoEncontrado
	}

	return produto, nil
}

func (m *mockProdutoRepository) FindAll() ([]domain.Produto, error) {

	produtos := make([]domain.Produto, 0, len(m.produtos))

	for _, produto := range m.produtos {
		produtos = append(produtos, *produto)
	}

	return produtos, nil
}

func (m *mockProdutoRepository) AjustarSaldo(
	codigo string,
	delta int,
) (*domain.Produto, error) {

	produto, exists := m.produtos[codigo]

	if !exists {
		return nil, repository.ErrProdutoNaoEncontrado
	}

	if err := produto.AjustarSaldo(delta); err != nil {
		return nil, err
	}

	return produto, nil
}

func TestProdutoService_CriarProduto(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	produto, err := svc.CriarProduto(
		"PROD-001",
		"Produto de teste",
		10,
	)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, "PROD-001", produto.Codigo)
	assert.Equal(t, "Produto de teste", produto.Descricao)
	assert.Equal(t, 10, produto.Saldo)
}

func TestProdutoService_CriarProdutoCodigoObrigatorio(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	produto, err := svc.CriarProduto(
		"",
		"Produto de teste",
		10,
	)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, service.ErrCodigoObrigatorio)
}

func TestProdutoService_CriarProdutoDescricaoObrigatoria(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	produto, err := svc.CriarProduto(
		"PROD-001",
		"",
		10,
	)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, service.ErrDescricaoObrigatoria)
}

func TestProdutoService_CriarProdutoSaldoInvalido(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	produto, err := svc.CriarProduto(
		"PROD-001",
		"Produto de teste",
		-1,
	)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, service.ErrSaldoInvalido)
}

func TestProdutoService_BuscarProduto(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	_, err := svc.CriarProduto(
		"PROD-001",
		"Produto de teste",
		10,
	)

	require.NoError(t, err)

	produto, err := svc.BuscarProduto("PROD-001")

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, "PROD-001", produto.Codigo)
	assert.Equal(t, 10, produto.Saldo)
}

func TestProdutoService_BuscarProdutoNaoEncontrado(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	produto, err := svc.BuscarProduto("PROD-999")

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, repository.ErrProdutoNaoEncontrado)
}

func TestProdutoService_ListarProdutos(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	require.NoError(t, func() error {
		_, err := svc.CriarProduto("PROD-001", "Produto 1", 10)
		return err
	}())

	require.NoError(t, func() error {
		_, err := svc.CriarProduto("PROD-002", "Produto 2", 20)
		return err
	}())

	produtos, err := svc.ListarProdutos()

	require.NoError(t, err)
	assert.Len(t, produtos, 2)
}

func TestProdutoService_AjustarSaldo(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	_, err := svc.CriarProduto(
		"PROD-001",
		"Produto de teste",
		10,
	)

	require.NoError(t, err)

	produto, err := svc.AjustarSaldo("PROD-001", -3)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, 7, produto.Saldo)
}

func TestProdutoService_AjustarSaldoIncremento(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	_, err := svc.CriarProduto(
		"PROD-001",
		"Produto de teste",
		10,
	)

	require.NoError(t, err)

	produto, err := svc.AjustarSaldo("PROD-001", 5)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, 15, produto.Saldo)
}

func TestProdutoService_AjustarSaldoInsuficiente(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	_, err := svc.CriarProduto(
		"PROD-001",
		"Produto de teste",
		5,
	)

	require.NoError(t, err)

	produto, err := svc.AjustarSaldo("PROD-001", -6)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, domain.ErrSaldoInsuficiente)
}

func TestProdutoService_AjustarSaldoZero(t *testing.T) {
	repo := newMockProdutoRepository()
	svc := service.NewProdutoService(repo)

	produto, err := svc.AjustarSaldo(
		"PROD-001",
		0,
	)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, service.ErrDeltaInvalido)
}
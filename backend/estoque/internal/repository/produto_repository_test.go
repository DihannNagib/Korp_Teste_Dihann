package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/config"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/database"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/repository"
)

func setupRepository(t *testing.T) repository.ProdutoRepository {
	t.Helper()

	cfg := config.Load("../../../../.env")
	db := database.Connect(cfg)

	cleanupTestData(t, db)

	t.Cleanup(func() {
		cleanupTestData(t, db)
	})

	return repository.NewProdutoRepository(db)
}

func cleanupTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	err := db.
		Where("codigo LIKE ?", "TESTE-%").
		Delete(&domain.Produto{}).
		Error

	require.NoError(t, err)
}

func TestProdutoRepository_Create(t *testing.T) {
	repo := setupRepository(t)

	err := repo.Create(&domain.Produto{
		Codigo:    "TESTE-CREATE-001",
		Descricao: "Produto de teste",
		Saldo:     10,
	})

	require.NoError(t, err)

	produto, err := repo.FindByCodigo("TESTE-CREATE-001")

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, "TESTE-CREATE-001", produto.Codigo)
	assert.Equal(t, "Produto de teste", produto.Descricao)
	assert.Equal(t, 10, produto.Saldo)
}

func TestProdutoRepository_CreateCodigoDuplicado(t *testing.T) {
	repo := setupRepository(t)

	codigo := "TESTE-DUPLICADO-001"

	err := repo.Create(&domain.Produto{
		Codigo:    codigo,
		Descricao: "Produto original",
		Saldo:     10,
	})

	require.NoError(t, err)

	err = repo.Create(&domain.Produto{
		Codigo:    codigo,
		Descricao: "Produto duplicado",
		Saldo:     20,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrCodigoJaExiste)
}

func TestProdutoRepository_FindByCodigo(t *testing.T) {
	repo := setupRepository(t)

	codigo := "TESTE-FIND-001"

	err := repo.Create(&domain.Produto{
		Codigo:    codigo,
		Descricao: "Produto para busca",
		Saldo:     15,
	})

	require.NoError(t, err)

	produto, err := repo.FindByCodigo(codigo)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, codigo, produto.Codigo)
	assert.Equal(t, "Produto para busca", produto.Descricao)
	assert.Equal(t, 15, produto.Saldo)
}

func TestProdutoRepository_FindByCodigoNaoEncontrado(t *testing.T) {
	repo := setupRepository(t)

	produto, err := repo.FindByCodigo("CODIGO-INEXISTENTE-999")

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, repository.ErrProdutoNaoEncontrado)
}

func TestProdutoRepository_FindAll(t *testing.T) {
	repo := setupRepository(t)

	require.NoError(t, repo.Create(&domain.Produto{
		Codigo:    "TESTE-FINDALL-002",
		Descricao: "Produto 2",
		Saldo:     20,
	}))

	require.NoError(t, repo.Create(&domain.Produto{
		Codigo:    "TESTE-FINDALL-001",
		Descricao: "Produto 1",
		Saldo:     10,
	}))

	produtos, err := repo.FindAll()

	require.NoError(t, err)
	require.Len(t, produtos, 2)

	assert.Equal(t, "TESTE-FINDALL-001", produtos[0].Codigo)
	assert.Equal(t, "TESTE-FINDALL-002", produtos[1].Codigo)
}

func TestProdutoRepository_AjustarSaldoReducao(t *testing.T) {
	repo := setupRepository(t)

	codigo := "TESTE-SALDO-001"

	require.NoError(t, repo.Create(&domain.Produto{
		Codigo:    codigo,
		Descricao: "Produto saldo",
		Saldo:     10,
	}))

	produto, err := repo.AjustarSaldo(codigo, -3)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, 7, produto.Saldo)
}

func TestProdutoRepository_AjustarSaldoIncremento(t *testing.T) {
	repo := setupRepository(t)

	codigo := "TESTE-SALDO-002"

	require.NoError(t, repo.Create(&domain.Produto{
		Codigo:    codigo,
		Descricao: "Produto entrada",
		Saldo:     10,
	}))

	produto, err := repo.AjustarSaldo(codigo, 5)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, 15, produto.Saldo)
}

func TestProdutoRepository_AjustarSaldoInsuficiente(t *testing.T) {
	repo := setupRepository(t)

	codigo := "TESTE-SALDO-003"

	require.NoError(t, repo.Create(&domain.Produto{
		Codigo:    codigo,
		Descricao: "Produto saldo insuficiente",
		Saldo:     5,
	}))

	produto, err := repo.AjustarSaldo(codigo, -6)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, domain.ErrSaldoInsuficiente)

	produto, err = repo.FindByCodigo(codigo)

	require.NoError(t, err)
	require.NotNil(t, produto)

	assert.Equal(t, 5, produto.Saldo)
}

func TestProdutoRepository_AjustarSaldoProdutoNaoEncontrado(t *testing.T) {
	repo := setupRepository(t)

	produto, err := repo.AjustarSaldo(
		"CODIGO-INEXISTENTE-999",
		-1,
	)

	assert.Nil(t, produto)
	assert.ErrorIs(t, err, repository.ErrProdutoNaoEncontrado)
}

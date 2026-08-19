package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/config"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/database"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
)

func setupRepository(t *testing.T) (repository.NotaFiscalRepository, *gorm.DB) {
	t.Helper()

	cfg := config.Load("../../../../.env")
	db := database.Connect(cfg)

	return repository.NewNotaFiscalRepository(db), db
}

func criarNotaComItens(
	t *testing.T,
	db *gorm.DB,
	itens ...struct {
		codigo     string
		quantidade int
	},
) *domain.NotaFiscal {
	t.Helper()

	nota := domain.NovaNotaFiscal()

	for _, item := range itens {
		require.NoError(
			t,
			nota.AdicionarItem(item.codigo, item.quantidade),
		)
	}

	repo := repository.NewNotaFiscalRepository(db)

	require.NoError(t, repo.Create(nota))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota.ID).
			Delete(&domain.NotaFiscal{})
	})

	return nota
}

func TestNotaFiscalRepository_CreateAtribuiNumeroSequencial(t *testing.T) {
	repo, db := setupRepository(t)

	nota1 := domain.NovaNotaFiscal()
	require.NoError(t, nota1.AdicionarItem("TESTE-SEQ-001", 1))
	require.NoError(t, repo.Create(nota1))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota1.ID).
			Delete(&domain.NotaFiscal{})
	})

	nota2 := domain.NovaNotaFiscal()
	require.NoError(t, nota2.AdicionarItem("TESTE-SEQ-002", 1))
	require.NoError(t, repo.Create(nota2))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota2.ID).
			Delete(&domain.NotaFiscal{})
	})

	assert.NotZero(t, nota1.Numero)
	assert.NotZero(t, nota2.Numero)
	assert.Greater(t, nota2.Numero, nota1.Numero)
}

func TestNotaFiscalRepository_CreatePersisteNotaEItens(t *testing.T) {
	repo, db := setupRepository(t)

	nota := domain.NovaNotaFiscal()

	require.NoError(
		t,
		nota.AdicionarItem("TESTE-ITEM-001", 3),
	)

	require.NoError(
		t,
		nota.AdicionarItem("TESTE-ITEM-002", 5),
	)

	require.NoError(t, repo.Create(nota))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota.ID).
			Delete(&domain.NotaFiscal{})
	})

	encontrada, err := repo.FindByNumero(nota.Numero)

	require.NoError(t, err)
	require.NotNil(t, encontrada)

	assert.Equal(t, nota.ID, encontrada.ID)
	assert.Equal(t, nota.Numero, encontrada.Numero)
	assert.Equal(t, domain.StatusAberta, encontrada.Status)

	require.Len(t, encontrada.Itens, 2)

	itens := map[string]int{}

	for _, item := range encontrada.Itens {
		itens[item.ProdutoCodigo] = item.Quantidade
	}

	assert.Equal(t, 3, itens["TESTE-ITEM-001"])
	assert.Equal(t, 5, itens["TESTE-ITEM-002"])
}

func TestNotaFiscalRepository_FindByNumeroNaoEncontrada(t *testing.T) {
	repo, _ := setupRepository(t)

	nota, err := repo.FindByNumero(999999999)

	assert.Nil(t, nota)
	assert.ErrorIs(t, err, repository.ErrNotaNaoEncontrada)
}

func TestNotaFiscalRepository_FindAll(t *testing.T) {
	repo, db := setupRepository(t)

	nota1 := domain.NovaNotaFiscal()
	require.NoError(t, nota1.AdicionarItem("TESTE-ALL-001", 1))
	require.NoError(t, repo.Create(nota1))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota1.ID).
			Delete(&domain.NotaFiscal{})
	})

	nota2 := domain.NovaNotaFiscal()
	require.NoError(t, nota2.AdicionarItem("TESTE-ALL-002", 1))
	require.NoError(t, repo.Create(nota2))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota2.ID).
			Delete(&domain.NotaFiscal{})
	})

	notas, err := repo.FindAll()

	require.NoError(t, err)

	var encontrouNota1 bool
	var encontrouNota2 bool

	for _, nota := range notas {
		switch nota.Numero {
		case nota1.Numero:
			encontrouNota1 = true

			require.Len(t, nota.Itens, 1)
			assert.Equal(t, "TESTE-ALL-001", nota.Itens[0].ProdutoCodigo)

		case nota2.Numero:
			encontrouNota2 = true

			require.Len(t, nota.Itens, 1)
			assert.Equal(t, "TESTE-ALL-002", nota.Itens[0].ProdutoCodigo)
		}
	}

	assert.True(t, encontrouNota1)
	assert.True(t, encontrouNota2)
}

func TestNotaFiscalRepository_AtualizarStatus(t *testing.T) {
	repo, db := setupRepository(t)

	nota := domain.NovaNotaFiscal()

	require.NoError(
		t,
		nota.AdicionarItem("TESTE-STATUS-001", 1),
	)

	require.NoError(t, repo.Create(nota))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota.ID).
			Delete(&domain.NotaFiscal{})
	})

	require.NoError(t, nota.Fechar())

	require.NoError(
		t,
		repo.AtualizarStatus(nota),
	)

	atualizada, err := repo.FindByNumero(nota.Numero)

	require.NoError(t, err)
	assert.Equal(t, domain.StatusFechada, atualizada.Status)
}

func TestNotaFiscalRepository_AtualizarStatusNaoAlteraItens(t *testing.T) {
	repo, db := setupRepository(t)

	nota := domain.NovaNotaFiscal()

	require.NoError(
		t,
		nota.AdicionarItem("TESTE-ITENS-PRESERVADOS", 7),
	)

	require.NoError(t, repo.Create(nota))

	t.Cleanup(func() {
		db.Unscoped().
			Where("id = ?", nota.ID).
			Delete(&domain.NotaFiscal{})
	})

	require.NoError(t, nota.Fechar())

	require.NoError(
		t,
		repo.AtualizarStatus(nota),
	)

	atualizada, err := repo.FindByNumero(nota.Numero)

	require.NoError(t, err)

	require.Len(t, atualizada.Itens, 1)

	assert.Equal(
		t,
		"TESTE-ITENS-PRESERVADOS",
		atualizada.Itens[0].ProdutoCodigo,
	)

	assert.Equal(
		t,
		7,
		atualizada.Itens[0].Quantidade,
	)

	assert.Equal(
		t,
		domain.StatusFechada,
		atualizada.Status,
	)
}

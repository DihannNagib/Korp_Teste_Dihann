package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
)

var (
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
  ErrCodigoJaExiste       = errors.New("código de produto já cadastrado")
)

type ProdutoRepository interface {
	Create(produto *domain.Produto) error
	FindByCodigo(codigo string) (*domain.Produto, error)
	FindAll() ([]domain.Produto, error)
	AjustarSaldo(codigo string, delta int) (*domain.Produto, error)
}

type produtoRepository struct {
	db *gorm.DB
}

func NewProdutoRepository(db *gorm.DB) ProdutoRepository {
	return &produtoRepository{
		db: db,
	}
}

func (r *produtoRepository) Create(produto *domain.Produto) error {
	err := r.db.Create(produto).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrCodigoJaExiste
	}

	return err
}

func (r *produtoRepository) FindByCodigo(codigo string) (*domain.Produto, error) {
	var produto domain.Produto

	err := r.db.
		Where("codigo = ?", codigo).
		First(&produto).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProdutoNaoEncontrado
	}

	if err != nil {
		return nil, err
	}

	return &produto, nil
}

func (r *produtoRepository) FindAll() ([]domain.Produto, error) {
	var produtos []domain.Produto

	err := r.db.
		Order("codigo").
		Find(&produtos).Error

	return produtos, err
}

func (r *produtoRepository) AjustarSaldo(
	codigo string,
	delta int,
) (*domain.Produto, error) {

	var produtoAtualizado domain.Produto

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var produto domain.Produto

		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("codigo = ?", codigo).
			First(&produto).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProdutoNaoEncontrado
		}

		if err != nil {
			return err
		}

		if err := produto.AjustarSaldo(delta); err != nil {
			return err
		}

		if err := tx.Save(&produto).Error; err != nil {
			return err
		}

		produtoAtualizado = produto

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &produtoAtualizado, nil
}
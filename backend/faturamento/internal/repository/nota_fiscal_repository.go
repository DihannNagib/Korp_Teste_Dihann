package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
)

var ErrNotaNaoEncontrada = errors.New("nota fiscal nao encontrada")

type NotaFiscalRepository interface {
	Create(nota *domain.NotaFiscal) error
	FindByNumero(numero uint) (*domain.NotaFiscal, error)
	FindAll() ([]domain.NotaFiscal, error)
	AtualizarStatus(nota *domain.NotaFiscal) error
}

type notaFiscalRepository struct {
	db *gorm.DB
}

func NewNotaFiscalRepository(db *gorm.DB) NotaFiscalRepository {
	return &notaFiscalRepository{
		db: db,
	}
}

func (r *notaFiscalRepository) Create(nota *domain.NotaFiscal) error {
	return r.db.Create(nota).Error
}

func (r *notaFiscalRepository) FindByNumero(numero uint) (*domain.NotaFiscal, error) {
	var nota domain.NotaFiscal

	err := r.db.
		Preload("Itens").
		Where("numero = ?", numero).
		First(&nota).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotaNaoEncontrada
	}

	if err != nil {
		return nil, err
	}

	return &nota, nil
}

func (r *notaFiscalRepository) FindAll() ([]domain.NotaFiscal, error) {
	var notas []domain.NotaFiscal

	err := r.db.
		Preload("Itens").
		Order("numero").
		Find(&notas).Error

	return notas, err
}

func (r *notaFiscalRepository) AtualizarStatus(nota *domain.NotaFiscal) error {
	return r.db.
		Model(&domain.NotaFiscal{}).
		Where("id = ?", nota.ID).
		Update("status", nota.Status).Error
}

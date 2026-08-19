package service

import (
	"errors"
	"fmt"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
)

var (
	ErrBaixaEstoque = errors.New("erro ao realizar baixa no estoque")
)

type EstoqueGateway interface {
	BaixarItens(itens []domain.ItemNotaFiscal) error
}

type NotaFiscalService interface {
	CriarNota(itens []ItemNotaFiscalInput) (*domain.NotaFiscal, error)
	BuscarPorNumero(numero uint) (*domain.NotaFiscal, error)
	Listar() ([]domain.NotaFiscal, error)
	Imprimir(numero uint) (*domain.NotaFiscal, error)
}

type notaFiscalService struct {
	repository repository.NotaFiscalRepository
	estoque    EstoqueGateway
}

type ItemNotaFiscalInput struct {
	ProdutoCodigo string
	Quantidade    int
}

func NewNotaFiscalService(
	repository repository.NotaFiscalRepository,
	estoque EstoqueGateway,
) NotaFiscalService {
	return &notaFiscalService{
		repository: repository,
		estoque:    estoque,
	}
}

func (s *notaFiscalService) CriarNota(
	itens []ItemNotaFiscalInput,
) (*domain.NotaFiscal, error) {
	nota := domain.NovaNotaFiscal()

	for _, item := range itens {
		if err := nota.AdicionarItem(
			item.ProdutoCodigo,
			item.Quantidade,
		); err != nil {
			return nil, err
		}
	}

	if err := s.repository.Create(nota); err != nil {
		return nil, fmt.Errorf("erro ao criar nota fiscal: %w", err)
	}

	return nota, nil
}

func (s *notaFiscalService) BuscarPorNumero(
	numero uint,
) (*domain.NotaFiscal, error) {
	return s.repository.FindByNumero(numero)
}

func (s *notaFiscalService) Listar() ([]domain.NotaFiscal, error) {
	return s.repository.FindAll()
}

func (s *notaFiscalService) Imprimir(
	numero uint,
) (*domain.NotaFiscal, error) {
	nota, err := s.repository.FindByNumero(numero)
	if err != nil {
		return nil, err
	}

	if err := nota.PodeSerImpressa(); err != nil {
		return nil, err
	}

	if err := s.estoque.BaixarItens(nota.Itens); err != nil {
		return nil, fmt.Errorf(
			"%w: %w",
			ErrBaixaEstoque,
			err,
		)
	}

	if err := nota.Fechar(); err != nil {
		return nil, err
	}

	if err := s.repository.AtualizarStatus(nota); err != nil {
		return nil, fmt.Errorf(
			"erro ao atualizar status da nota fiscal: %w",
			err,
		)
	}

	return nota, nil
}
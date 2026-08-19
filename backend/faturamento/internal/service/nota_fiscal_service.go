package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/client"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/repository"
)

var (
	ErrFalhaComunicacaoEstoque     = errors.New("falha ao comunicar com o servico de estoque")
	ErrSaldoInsuficienteEstoque    = errors.New("saldo insuficiente no estoque")
	ErrProdutoNaoEncontradoEstoque = errors.New("produto nao encontrado no estoque")
)

// EstoqueGateway e a interface que o service depende, permitindo mock
// em testes unitarios sem subir o servico de Estoque real.
type EstoqueGateway interface {
	BaixarItem(produtoCodigo string, quantidade int) error
	EstornarItem(produtoCodigo string, quantidade int) error
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

func NewNotaFiscalService(repository repository.NotaFiscalRepository, estoque EstoqueGateway) NotaFiscalService {
	return &notaFiscalService{repository: repository, estoque: estoque}
}

func (s *notaFiscalService) CriarNota(itens []ItemNotaFiscalInput) (*domain.NotaFiscal, error) {
	nota := domain.NovaNotaFiscal()
	for _, item := range itens {
		if err := nota.AdicionarItem(item.ProdutoCodigo, item.Quantidade); err != nil {
			return nil, err
		}
	}
	if err := s.repository.Create(nota); err != nil {
		return nil, fmt.Errorf("erro ao criar nota fiscal: %w", err)
	}
	return nota, nil
}

func (s *notaFiscalService) BuscarPorNumero(numero uint) (*domain.NotaFiscal, error) {
	return s.repository.FindByNumero(numero)
}

func (s *notaFiscalService) Listar() ([]domain.NotaFiscal, error) {
	return s.repository.FindAll()
}

// Imprimir baixa o saldo de cada item, um a um. Se algum item falhar,
// todos os itens ja baixados com sucesso nesta tentativa sao
// compensados (estornados) antes do erro ser propagado -- a nota
// permanece ABERTA e nenhum saldo fica baixado "orfao".
func (s *notaFiscalService) Imprimir(numero uint) (*domain.NotaFiscal, error) {
	nota, err := s.repository.FindByNumero(numero)
	if err != nil {
		return nil, err
	}

	if err := nota.PodeSerImpressa(); err != nil {
		return nil, err
	}

	processados := make([]domain.ItemNotaFiscal, 0, len(nota.Itens))

	for _, item := range nota.Itens {
		if err := s.estoque.BaixarItem(item.ProdutoCodigo, item.Quantidade); err != nil {
			s.compensar(processados)
			return nil, mapearErroEstoque(err, item.ProdutoCodigo)
		}
		processados = append(processados, item)
	}

	if err := nota.Fechar(); err != nil {
		s.compensar(processados)
		return nil, err
	}

	if err := s.repository.AtualizarStatus(nota); err != nil {
		s.compensar(processados)
		return nil, fmt.Errorf("erro ao atualizar status da nota fiscal: %w", err)
	}

	return nota, nil
}

// compensar e "melhor esforco": se o proprio estorno falhar (ex:
// Estoque caiu no meio da compensacao), registramos um alerta em log
// para reconciliacao manual em vez de perder o erro silenciosamente.
func (s *notaFiscalService) compensar(itens []domain.ItemNotaFiscal) {
	for _, item := range itens {
		if err := s.estoque.EstornarItem(item.ProdutoCodigo, item.Quantidade); err != nil {
			log.Printf(
				"[FATURAMENTO] ALERTA: falha ao compensar saldo do produto %s (quantidade %d): %v",
				item.ProdutoCodigo, item.Quantidade, err,
			)
		}
	}
}

func mapearErroEstoque(err error, codigo string) error {
	switch {
	case errors.Is(err, client.ErrSaldoInsuficiente):
		return fmt.Errorf("%w (produto %s)", ErrSaldoInsuficienteEstoque, codigo)
	case errors.Is(err, client.ErrProdutoNaoEncontrado):
		return fmt.Errorf("%w (produto %s)", ErrProdutoNaoEncontradoEstoque, codigo)
	default:
		return fmt.Errorf("%w (produto %s): %v", ErrFalhaComunicacaoEstoque, codigo, err)
	}
}
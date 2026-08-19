package service

import (
	"errors"
	"strings"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/domain"
	"github.com/DihannNagib/Korp_Teste_Dihann/backend/estoque/internal/repository"
)

var (
	ErrCodigoObrigatorio    = errors.New("código obrigatório")
  ErrDescricaoObrigatoria = errors.New("descrição obrigatória")
  ErrSaldoInvalido        = errors.New("saldo inicial não pode ser negativo")
  ErrDeltaInvalido        = errors.New("delta não pode ser zero")
)

type ProdutoService interface {
	CriarProduto(codigo, descricao string, saldo int) (*domain.Produto, error)
	BuscarProduto(codigo string) (*domain.Produto, error)
	ListarProdutos() ([]domain.Produto, error)
	AjustarSaldo(codigo string, delta int) (*domain.Produto, error)
}

type produtoService struct {
	repository repository.ProdutoRepository
}

func NewProdutoService(repository repository.ProdutoRepository) ProdutoService {
	return &produtoService{
		repository: repository,
	}
}

func (s *produtoService) CriarProduto(
	codigo string,
	descricao string,
	saldo int,
) (*domain.Produto, error) {

	codigo = strings.TrimSpace(codigo)
	descricao = strings.TrimSpace(descricao)

	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}

	if descricao == "" {
		return nil, ErrDescricaoObrigatoria
	}

	if saldo < 0 {
		return nil, ErrSaldoInvalido
	}

	produto := &domain.Produto{
		Codigo:    codigo,
		Descricao: descricao,
		Saldo:     saldo,
	}

	if err := s.repository.Create(produto); err != nil {
		return nil, err
	}

	return produto, nil
}

func (s *produtoService) BuscarProduto(
	codigo string,
) (*domain.Produto, error) {

	codigo = strings.TrimSpace(codigo)

	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}

	return s.repository.FindByCodigo(codigo)
}

func (s *produtoService) ListarProdutos() ([]domain.Produto, error) {
	return s.repository.FindAll()
}

func (s *produtoService) AjustarSaldo(
	codigo string,
	delta int,
) (*domain.Produto, error) {

	codigo = strings.TrimSpace(codigo)

	if codigo == "" {
		return nil, ErrCodigoObrigatorio
	}

	if delta == 0 {
		return nil, ErrDeltaInvalido
	}

	return s.repository.AjustarSaldo(codigo, delta)
}
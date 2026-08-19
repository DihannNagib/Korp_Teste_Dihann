package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DihannNagib/Korp_Teste_Dihann/backend/faturamento/internal/domain"
)

var (
	ErrEstoqueIndisponivel = errors.New("serviço de estoque indisponível")
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado no estoque")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente no estoque")
)

type EstoqueClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewEstoqueClient(baseURL string) *EstoqueClient {
	return &EstoqueClient{
		baseURL: baseURL,
		httpClient: &http.Client{},
	}
}

type ajustarSaldoRequest struct {
	Delta int `json:"delta"`
}

func (c *EstoqueClient) BaixarItens(
	itens []domain.ItemNotaFiscal,
) error {
	for _, item := range itens {
		if err := c.BaixarItem(
			item.ProdutoCodigo,
			item.Quantidade,
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *EstoqueClient) BaixarItem(
	produtoCodigo string,
	quantidade int,
) error {
	payload := ajustarSaldoRequest{
		Delta: -quantidade,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar baixa de estoque: %w", err)
	}

	url := fmt.Sprintf(
		"%s/api/v1/produtos/%s/saldo",
		c.baseURL,
		produtoCodigo,
	)

	req, err := http.NewRequest(
		http.MethodPatch,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição para estoque: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"%w: %v",
			ErrEstoqueIndisponivel,
			err,
		)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	case http.StatusNotFound:
		return ErrProdutoNaoEncontrado

	case http.StatusUnprocessableEntity:
		return ErrSaldoInsuficiente

	default:
		return fmt.Errorf(
			"estoque retornou status HTTP %d",
			resp.StatusCode,
		)
	}
}
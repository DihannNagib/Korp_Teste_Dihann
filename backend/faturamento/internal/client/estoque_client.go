package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrEstoqueIndisponivel = errors.New("servico de estoque indisponivel")
	ErrProdutoNaoEncontrado = errors.New("produto nao encontrado no estoque")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente no estoque")
)

type EstoqueClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewEstoqueClient(baseURL string) *EstoqueClient {
	return &EstoqueClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type ajustarSaldoRequest struct {
	Delta int `json:"delta"`
}

func (c *EstoqueClient) BaixarItem(produtoCodigo string, quantidade int) error {
	return c.ajustarSaldo(produtoCodigo, -quantidade)
}

func (c *EstoqueClient) EstornarItem(produtoCodigo string, quantidade int) error {
	return c.ajustarSaldo(produtoCodigo, quantidade)
}

func (c *EstoqueClient) ajustarSaldo(produtoCodigo string, delta int) error {
	payload := ajustarSaldoRequest{Delta: delta}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar ajuste de estoque: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/produtos/%s/saldo", c.baseURL, produtoCodigo)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("erro ao criar requisicao para estoque: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEstoqueIndisponivel, err)
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
		return fmt.Errorf("%w: estoque retornou status HTTP %d", ErrEstoqueIndisponivel, resp.StatusCode)
	}
}
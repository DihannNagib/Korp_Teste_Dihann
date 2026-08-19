package dto

type CriarProdutoRequest struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

type AjustarSaldoRequest struct {
	Delta int `json:"delta"`
}
package dto

type ItemRequest struct {
	ProdutoCodigo string `json:"produtoCodigo" binding:"required"`
	Quantidade int `json:"quantidade" binding:"required,gt=0"`
}

type CriarNotaRequest struct {
	Itens []ItemRequest `json:"itens" binding:"required,min=1,dive"`
}
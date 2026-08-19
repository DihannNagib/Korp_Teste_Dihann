package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusAberta = "ABERTA"
	StatusFechada = "FECHADA"
)

var (
	ErrNotaNaoAberta = errors.New("nota fiscal não está aberta")
	ErrQuantidadeInvalida = errors.New("quantidade deve ser maior que zero")
	ErrProdutoInvalido = errors.New("código do produto é obrigatório")
	ErrSemItens = errors.New("nota fiscal deve conter ao menos um item")
)

type ItemNotaFiscal struct {
	ID uint `gorm:"primaryKey" json:"id"`
	NotaFiscalID uint `gorm:"not null" json:"notaFiscalId"`
	ProdutoCodigo string `gorm:"not null" json:"produtoCodigo"`
	Quantidade int  `gorm:"not null" json:"quantidade"`
}

func (ItemNotaFiscal) TableName() string {
	return "itens_nota_fiscal"
}

type NotaFiscal struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Numero uint `gorm:"unique;not null;default:nextval('notas_fiscais_numero_seq')" json:"numero"`
	Status string `gorm:"not null" json:"status"`
	Itens []ItemNotaFiscal `gorm:"foreignKey:NotaFiscalID" json:"itens"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (NotaFiscal) TableName() string {
	return "notas_fiscais"
}

func NovaNotaFiscal() *NotaFiscal {
	return &NotaFiscal{
		Status: StatusAberta,
		Itens:  make([]ItemNotaFiscal, 0),
	}
}

func (n *NotaFiscal) AdicionarItem(
	produtoCodigo string,
	quantidade int,
) error {
	if n.Status != StatusAberta {
		return ErrNotaNaoAberta
	}

	produtoCodigo = strings.TrimSpace(produtoCodigo)

	if produtoCodigo == "" {
			return ErrProdutoInvalido
	}

	if quantidade <= 0 {
		return ErrQuantidadeInvalida
	}

	n.Itens = append(n.Itens, ItemNotaFiscal{
		ProdutoCodigo: produtoCodigo,
		Quantidade:    quantidade,
	})

	return nil
}

func (n *NotaFiscal) PodeSerImpressa() error {
	if n.Status != StatusAberta {
		return ErrNotaNaoAberta
	}

	if len(n.Itens) == 0 {
		return ErrSemItens
	}

	return nil
}

func (n *NotaFiscal) Fechar() error {
	if n.Status != StatusAberta {
		return ErrNotaNaoAberta
	}

	n.Status = StatusFechada

	return nil
}
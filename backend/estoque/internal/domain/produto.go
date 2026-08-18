package domain

import (
	"errors"
	"time"
)

var ErrSaldoInsuficiente = errors.New("saldo insuficiente")

type Produto struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Codigo string `gorm:"uniqueIndex;not null" json:"codigo"`
	Descricao string `gorm:"not null" json:"descricao"`
	Saldo int `gorm:"not null" json:"saldo"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (p *Produto) AjustarSaldo(delta int) error {
	novoSaldo := p.Saldo + delta

	if novoSaldo < 0 {
		return ErrSaldoInsuficiente
	}

	p.Saldo = novoSaldo

	return nil
}
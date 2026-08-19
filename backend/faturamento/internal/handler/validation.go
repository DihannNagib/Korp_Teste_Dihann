package handler

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type CampoErro struct {
	Campo string `json:"campo"`
	Erro  string `json:"erro"`
}

func traduzirErroValidacao(
	ve validator.ValidationErrors,
) []CampoErro {
	erros := make([]CampoErro, 0, len(ve))

	for _, fe := range ve {
		erros = append(erros, CampoErro{
			Campo: fe.Field(),
			Erro:  mensagemPorTag(fe),
		})
	}

	return erros
}

func mensagemPorTag(fe validator.FieldError) string {
	campo := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s é obrigatório", campo)

	case "min":
		return fmt.Sprintf(
			"%s deve conter ao menos %s item(ns)",
			campo,
			fe.Param(),
		)

	case "gt":
		return fmt.Sprintf(
			"%s deve ser maior que %s",
			campo,
			fe.Param(),
		)

	default:
		return fmt.Sprintf("%s inválido", campo)
	}
}
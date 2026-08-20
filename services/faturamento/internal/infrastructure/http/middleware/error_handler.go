package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	domainerrors "faturamento-service/internal/domain/errors"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var code int
			var message string

			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				code = http.StatusBadRequest
				var msgs []string
				for _, fe := range ve {
					msgs = append(msgs, formatFieldError(fe))
				}
				message = strings.Join(msgs, "; ")
				c.JSON(code, gin.H{
					"error": AppError{
						Code:    code,
						Message: message,
					},
				})
				return
			}

			switch {
			case errors.Is(err, domainerrors.ErrNotaFiscalNotFound) ||
				errors.Is(err, domainerrors.ErrProdutoNaoEncontrado):
				code = http.StatusNotFound
				message = err.Error()
			case errors.Is(err, domainerrors.ErrNotaFiscalNaoAberta) ||
				errors.Is(err, domainerrors.ErrConcurrencyConflict) ||
				errors.Is(err, domainerrors.ErrSaldoInsuficiente):
				code = http.StatusConflict
				message = err.Error()
			case errors.Is(err, domainerrors.ErrItensObrigatorios) ||
				errors.Is(err, domainerrors.ErrQuantidadeInvalida) ||
				errors.Is(err, domainerrors.ErrProdutoIDObrigatorio):
				code = http.StatusUnprocessableEntity
				message = err.Error()
			case errors.Is(err, domainerrors.ErrEstoqueIndisponivel):
				code = http.StatusServiceUnavailable
				message = err.Error()
			case strings.HasPrefix(err.Error(), "estoque:"):
				code = http.StatusBadGateway
				message = err.Error()
			default:
				code = http.StatusInternalServerError
				message = "Erro interno do servidor"
			}

			if code >= http.StatusInternalServerError {
				log.Printf("[faturamento] %s %s -> %d: %v", c.Request.Method, c.Request.URL.Path, code, err)
			}

			c.JSON(code, gin.H{
				"error": AppError{
					Code:    code,
					Message: message,
				},
			})
		}
	}
}

func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " é obrigatório"
	case "gt":
		return field + " deve ser maior que " + fe.Param()
	case "min":
		return field + " deve ter no mínimo " + fe.Param() + " itens"
	default:
		return field + " é inválido"
	}
}

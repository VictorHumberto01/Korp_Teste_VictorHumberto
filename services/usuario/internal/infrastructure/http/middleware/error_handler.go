package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	domainerrors "usuario-service/internal/domain/errors"
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

			// Check for validation errors from ShouldBindJSON
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
			case errors.Is(err, domainerrors.ErrUsuarioNotFound):
				code = http.StatusNotFound
				message = err.Error()
			case errors.Is(err, domainerrors.ErrEmailAlreadyExists) ||
				errors.Is(err, domainerrors.ErrCPFAlreadyExists) ||
				errors.Is(err, domainerrors.ErrConcurrencyConflict):
				code = http.StatusConflict
				message = err.Error()
			case errors.Is(err, domainerrors.ErrCPFInvalido) ||
				errors.Is(err, domainerrors.ErrEmailInvalido) ||
				errors.Is(err, domainerrors.ErrNomeInvalido) ||
				errors.Is(err, domainerrors.ErrUsuarioInativo) ||
				errors.Is(err, domainerrors.ErrUsuarioAtivo):
				code = http.StatusUnprocessableEntity
				message = err.Error()
			default:
				code = http.StatusInternalServerError
				message = "Erro interno do servidor"
			}

			if code >= http.StatusInternalServerError {
				log.Printf("[usuario] %s %s -> %d: %v", c.Request.Method, c.Request.URL.Path, code, err)
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
	case "email":
		return field + " deve ser um email válido"
	case "min":
		return field + " deve ter no mínimo " + fe.Param() + " caracteres"
	case "max":
		return field + " deve ter no máximo " + fe.Param() + " caracteres"
	default:
		return field + " é inválido"
	}
}

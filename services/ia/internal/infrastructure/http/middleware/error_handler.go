package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, formatFieldError(fe))
			}
			message := strings.Join(msgs, "; ")
			log.Printf("[ia] %s %s -> 400: %s", c.Request.Method, c.Request.URL.Path, message)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": AppError{Code: http.StatusBadRequest, Message: message},
			})
			return
		}

		code := http.StatusBadGateway
		message := "Não foi possível gerar o conteúdo de IA no momento"

		log.Printf("[ia] %s %s -> %d: %v", c.Request.Method, c.Request.URL.Path, code, err)

		c.JSON(code, gin.H{
			"error": AppError{Code: code, Message: message},
		})
	}
}

func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " é obrigatório"
	case "min":
		return field + " deve ter no mínimo " + fe.Param() + " caracteres"
	case "max":
		return field + " deve ter no máximo " + fe.Param() + " caracteres"
	default:
		return field + " é inválido"
	}
}

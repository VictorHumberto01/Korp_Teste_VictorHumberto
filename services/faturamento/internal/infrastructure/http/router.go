package http

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"faturamento-service/internal/infrastructure/http/handler"
	"faturamento-service/internal/infrastructure/http/middleware"
)

func NewRouter(h *handler.NotaFiscalHandler, db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:4200"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
	}))

	r.Use(middleware.ErrorHandler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		notas := api.Group("/notas-fiscais")
		{
			notas.GET("", h.List)
			notas.GET("/:id", h.GetByID)

			// Rotas com idempotência (escrita)
			writeGroup := notas.Group("")
			writeGroup.Use(middleware.IdempotencyMiddleware(db))
			{
				writeGroup.POST("", h.Create)
				writeGroup.POST("/:id/imprimir", h.Imprimir)
			}
		}
	}

	return r
}

package http

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"estoque-service/internal/infrastructure/http/handler"
	"estoque-service/internal/infrastructure/http/middleware"
)

func NewRouter(h *handler.ProdutoHandler, db *gorm.DB) *gin.Engine {
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
		produtos := api.Group("/produtos")
		{
			produtos.GET("", h.List)
			produtos.GET("/:id", h.GetByID)
			produtos.DELETE("/:id", h.Delete)
			produtos.POST("/suggest-description", h.SuggestDescription)

			// Rotas com idempotência (escrita)
			writeGroup := produtos.Group("")
			writeGroup.Use(middleware.IdempotencyMiddleware(db))
			{
				writeGroup.POST("", h.Create)
				writeGroup.PUT("/:id", h.Update)
				writeGroup.POST("/:id/saldo/debitar", h.DebitarSaldo)
				writeGroup.POST("/:id/saldo/creditar", h.CreditarSaldo)
			}
		}
	}

	return r
}

package http

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"usuario-service/internal/infrastructure/http/handler"
	"usuario-service/internal/infrastructure/http/middleware"
)

func NewRouter(h *handler.UsuarioHandler, db *gorm.DB) *gin.Engine {
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
		usuarios := api.Group("/usuarios")
		{
			usuarios.GET("", h.List)
			usuarios.GET("/:id", h.GetByID)
			usuarios.DELETE("/:id", h.Delete)
			usuarios.POST("/suggest-bio", h.SuggestBio)

			// Routes with idempotency
			writeGroup := usuarios.Group("")
			writeGroup.Use(middleware.IdempotencyMiddleware(db))
			{
				writeGroup.POST("", h.Create)
				writeGroup.PUT("/:id", h.Update)
			}
		}
	}

	return r
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ia-service/internal/application/dto"
	"ia-service/internal/application/query"
)

type IAHandler struct {
	gerarDescricaoProdutoQuery *query.GerarDescricaoProdutoHandler
}

func NewIAHandler(gerarDescricaoProdutoQuery *query.GerarDescricaoProdutoHandler) *IAHandler {
	return &IAHandler{gerarDescricaoProdutoQuery: gerarDescricaoProdutoQuery}
}

func (h *IAHandler) GerarDescricaoProduto(c *gin.Context) {
	var req dto.GerarDescricaoProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.gerarDescricaoProdutoQuery.Handle(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

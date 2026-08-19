package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"estoque-service/internal/application/command"
	"estoque-service/internal/application/dto"
	"estoque-service/internal/application/query"
)

type ProdutoHandler struct {
	createCmd   *command.CreateProdutoHandler
	updateCmd   *command.UpdateProdutoHandler
	deleteCmd   *command.DeleteProdutoHandler
	debitarCmd  *command.DebitarSaldoHandler
	creditarCmd *command.CreditarSaldoHandler
	getQuery    *query.GetProdutoHandler
	listQuery   *query.ListProdutosHandler
}

func NewProdutoHandler(
	createCmd *command.CreateProdutoHandler,
	updateCmd *command.UpdateProdutoHandler,
	deleteCmd *command.DeleteProdutoHandler,
	debitarCmd *command.DebitarSaldoHandler,
	creditarCmd *command.CreditarSaldoHandler,
	getQuery *query.GetProdutoHandler,
	listQuery *query.ListProdutosHandler,
) *ProdutoHandler {
	return &ProdutoHandler{
		createCmd:   createCmd,
		updateCmd:   updateCmd,
		deleteCmd:   deleteCmd,
		debitarCmd:  debitarCmd,
		creditarCmd: creditarCmd,
		getQuery:    getQuery,
		listQuery:   listQuery,
	}
}

func (h *ProdutoHandler) Create(c *gin.Context) {
	var req dto.CreateProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.createCmd.Handle(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ProdutoHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getQuery.Handle(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProdutoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	resp, err := h.listQuery.Handle(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProdutoHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.updateCmd.Handle(c.Request.Context(), id, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProdutoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.deleteCmd.Handle(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProdutoHandler) DebitarSaldo(c *gin.Context) {
	id := c.Param("id")
	var req dto.SaldoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.debitarCmd.Handle(c.Request.Context(), id, req.Quantidade)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProdutoHandler) CreditarSaldo(c *gin.Context) {
	id := c.Param("id")
	var req dto.SaldoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.creditarCmd.Handle(c.Request.Context(), id, req.Quantidade)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

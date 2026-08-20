package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"faturamento-service/internal/application/command"
	"faturamento-service/internal/application/dto"
	"faturamento-service/internal/application/query"
)

type NotaFiscalHandler struct {
	createCmd   *command.CreateNotaFiscalHandler
	imprimirCmd *command.ImprimirNotaFiscalHandler
	getQuery    *query.GetNotaFiscalHandler
	listQuery   *query.ListNotasFiscaisHandler
}

func NewNotaFiscalHandler(
	createCmd *command.CreateNotaFiscalHandler,
	imprimirCmd *command.ImprimirNotaFiscalHandler,
	getQuery *query.GetNotaFiscalHandler,
	listQuery *query.ListNotasFiscaisHandler,
) *NotaFiscalHandler {
	return &NotaFiscalHandler{
		createCmd:   createCmd,
		imprimirCmd: imprimirCmd,
		getQuery:    getQuery,
		listQuery:   listQuery,
	}
}

func (h *NotaFiscalHandler) Create(c *gin.Context) {
	var req dto.CreateNotaFiscalRequest
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

func (h *NotaFiscalHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getQuery.Handle(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *NotaFiscalHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	resp, err := h.listQuery.Handle(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *NotaFiscalHandler) Imprimir(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.imprimirCmd.Handle(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"usuario-service/internal/application/command"
	"usuario-service/internal/application/dto"
	"usuario-service/internal/application/query"
)

type UsuarioHandler struct {
	createCmd       *command.CreateUsuarioHandler
	updateCmd       *command.UpdateUsuarioHandler
	deleteCmd       *command.DeleteUsuarioHandler
	getQuery        *query.GetUsuarioHandler
	listQuery       *query.ListUsuariosHandler
	suggestBioQuery *query.SuggestBioHandler
}

func NewUsuarioHandler(
	createCmd *command.CreateUsuarioHandler,
	updateCmd *command.UpdateUsuarioHandler,
	deleteCmd *command.DeleteUsuarioHandler,
	getQuery *query.GetUsuarioHandler,
	listQuery *query.ListUsuariosHandler,
	suggestBioQuery *query.SuggestBioHandler,
) *UsuarioHandler {
	return &UsuarioHandler{
		createCmd:       createCmd,
		updateCmd:       updateCmd,
		deleteCmd:       deleteCmd,
		getQuery:        getQuery,
		listQuery:       listQuery,
		suggestBioQuery: suggestBioQuery,
	}
}

func (h *UsuarioHandler) Create(c *gin.Context) {
	var req dto.CreateUsuarioRequest
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

func (h *UsuarioHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getQuery.Handle(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UsuarioHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	resp, err := h.listQuery.Handle(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UsuarioHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUsuarioRequest
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

func (h *UsuarioHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.deleteCmd.Handle(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UsuarioHandler) SuggestBio(c *gin.Context) {
	var req dto.SuggestBioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.suggestBioQuery.Handle(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

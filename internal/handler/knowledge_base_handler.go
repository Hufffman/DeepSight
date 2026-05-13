package handler

import (
	"errors"
	"net/http"
	"strconv"

	"DeepSight/internal/dto"
	"DeepSight/internal/middleware"
	"DeepSight/internal/service"

	"github.com/gin-gonic/gin"
)

type KnowledgeBaseHandler struct {
	kbService *service.KnowledgeBaseService
}

func NewKnowledgeBaseHandler(kbService *service.KnowledgeBaseService) *KnowledgeBaseHandler {
	return &KnowledgeBaseHandler{
		kbService: kbService,
	}
}

func (h *KnowledgeBaseHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("", h.CreateKnowledgeBase)
	r.GET("", h.ListKnowledgeBases)
	r.GET("/:id", h.GetKnowledgeBase)
	r.PUT("/:id", h.UpdateKnowledgeBase)
	r.DELETE("/:id", h.DeleteKnowledgeBase)
}

func (h *KnowledgeBaseHandler) CreateKnowledgeBase(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kb, err := h.kbService.CreateKnowledgeBase(userID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, service.ToKBResponse(kb, 0))
}

func (h *KnowledgeBaseHandler) ListKnowledgeBases(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.kbService.ListKnowledgeBaseResponses(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *KnowledgeBaseHandler) GetKnowledgeBase(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	response, err := h.kbService.GetKnowledgeBaseDetailForUser(userID, uint(id))
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge base not found"})
		case errors.Is(err, dto.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "knowledge base not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *KnowledgeBaseHandler) UpdateKnowledgeBase(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req dto.UpdateKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.kbService.UpdateKnowledgeBaseForUser(userID, uint(id), req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge base not found"})
		case errors.Is(err, dto.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "knowledge base not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *KnowledgeBaseHandler) DeleteKnowledgeBase(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.kbService.DeleteKnowledgeBaseForUser(userID, uint(id)); err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge base not found"})
		case errors.Is(err, dto.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "knowledge base not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "knowledge base deleted successfully"})
}

package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"DeepSight/internal/dto"
	"DeepSight/internal/middleware"
	"DeepSight/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("", h.CreateConversation)
	r.GET("", h.ListConversations)
	r.GET("/:id", h.GetConversation)
	r.DELETE("/:id", h.DeleteConversation)
	r.POST("/:id/chat", h.Chat)
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.chatService.CreateConversation(userID, req.KnowledgeBaseID, req.Title)
	if err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge base not found"})
		} else if errors.Is(err, dto.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "knowledge base not owned by user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, service.ToConversationResponse(conv))
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.chatService.ListConversations(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.chatService.GetConversation(userID, uint(id), page, pageSize)
	if err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		} else if errors.Is(err, dto.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "conversation not owned by user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.chatService.DeleteConversation(userID, uint(id)); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		} else if errors.Is(err, dto.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "conversation not owned by user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted successfully"})
}

func (h *ChatHandler) Chat(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 先设置所有 SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // nginx

	// 立即写入 header（不等待 body）
	c.Writer.WriteHeader(http.StatusOK)

	// 强制 flush headers
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}

	// 创建 context，支持客户端断开连接
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// 获取流式 channel
	stream, resultCallback, err := h.chatService.ChatStream(ctx, userID, uint(id), req.Question)
	if err != nil {
		var errMsg string
		if errors.Is(err, dto.ErrNotFound) {
			errMsg = "conversation not found"
		} else if errors.Is(err, dto.ErrForbidden) {
			errMsg = "conversation not owned by user"
		} else {
			errMsg = err.Error()
		}
		// SSE 格式发送错误
		c.Writer.WriteString(fmt.Sprintf("data: [ERROR] %s\n\n", errMsg))
		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	// 流式发送每个 token
	for {
		select {
		case <-ctx.Done():
			// 客户端断开
			return
		case token, ok := <-stream:
			if !ok {
				// stream 关闭，发送完成事件
				_, err := resultCallback()
				if err != nil {
					c.Writer.WriteString(fmt.Sprintf("data: [ERROR] %s\n\n", err.Error()))
				} else {
					c.Writer.WriteString(fmt.Sprintf("data: [DONE]\n\n"))
				}
				if f, ok := c.Writer.(http.Flusher); ok {
					f.Flush()
				}
				return
			}
			// 写入并立即 flush
			c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", token))
			if f, ok := c.Writer.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

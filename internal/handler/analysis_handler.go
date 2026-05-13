package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"DeepSight/internal/dto"
	"DeepSight/internal/middleware"
	"DeepSight/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct {
	analysisService *service.AnalysisService
}

func NewAnalysisHandler(analysisService *service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{analysisService: analysisService}
}

// RegisterRoutes registers analysis-related routes on the given router group.
func (h *AnalysisHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/:id", h.StreamDeepAnalysis)
	r.GET("/reports", h.ListReports)
	r.GET("/reports/:id", h.GetReport)
	r.DELETE("/reports/:id", h.DeleteReport)
}

// ListReports returns paginated analysis reports for the current user.
func (h *AnalysisHandler) ListReports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "200"))

	response, err := h.analysisService.ListAllReports(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetReport returns a single analysis report by ID.
func (h *AnalysisHandler) GetReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	report, err := h.analysisService.GetReport(userID, uint(reportID))
	if err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// DeleteReport deletes an analysis report after ownership verification.
func (h *AnalysisHandler) DeleteReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	if err := h.analysisService.DeleteReport(userID, uint(reportID)); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return
		}
		if errors.Is(err, dto.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "report deleted"})
}

// StreamDeepAnalysis SSE streams deep analysis progress and results.
func (h *AnalysisHandler) StreamDeepAnalysis(c *gin.Context) {
	userID := middleware.GetUserID(c)

	kbID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid knowledge base id"})
		return
	}

	convID, _ := strconv.ParseUint(c.DefaultQuery("conv_id", "0"), 10, 32)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	// Heartbeat to prevent HTTP/2 stream timeout from proxies (Traefik/Nginx)
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.Writer.WriteString(": heartbeat\n\n")
				if flusher, ok := c.Writer.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-done:
				return
			}
		}
	}()

	writeSSE := func(event dto.AnalysisSSEEvent) {
		data, _ := json.Marshal(event)
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(data)))
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	events, err := h.analysisService.StreamAnalysis(c.Request.Context(), userID, uint(kbID), uint(convID))
	if err != nil {
		writeSSE(dto.AnalysisSSEEvent{Type: "error", Content: err.Error()})
		return
	}

	for event := range events {
		writeSSE(event)
	}
}

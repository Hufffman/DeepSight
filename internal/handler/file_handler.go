package handler

import (
	errors2 "DeepSight/internal/dto"
	"errors"
	"net/http"
	"strconv"

	"DeepSight/internal/middleware"
	"DeepSight/internal/service"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) RegisterRoutes(kbs *gin.RouterGroup) {
	kbs.POST("/:id/files", h.UploadFile)
	kbs.GET("/:id/files", h.ListFiles)
	kbs.GET("/:id/files/:file_id", h.GetFile)
	kbs.DELETE("/:id/files/:file_id", h.DeleteFile)
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	kbID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	response, err := h.fileService.UploadFileForUser(userID, uint(kbID), fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		switch {
		case errors.Is(err, errors2.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge base not found"})
		case errors.Is(err, errors2.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "knowledge base not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	userID := middleware.GetUserID(c)

	kbID, err := strconv.ParseUint(c.Param("id"), 10, 32)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.fileService.ListFilesForUser(userID, uint(kbID), page, pageSize)
	if err != nil {
		switch {
		case errors.Is(err, errors2.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge base not found"})
		case errors.Is(err, errors2.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "knowledge base not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *FileHandler) GetFile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	kbID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	fileID, _ := strconv.ParseUint(c.Param("file_id"), 10, 32)

	response, err := h.fileService.GetFileDetailForUser(userID, uint(kbID), uint(fileID))
	if err != nil {
		switch {
		case errors.Is(err, errors2.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "file or knowledge base not found"})
		case errors.Is(err, errors2.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "resource not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	kbID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	fileID, _ := strconv.ParseUint(c.Param("file_id"), 10, 32)

	if err := h.fileService.DeleteFileForUser(userID, uint(kbID), uint(fileID)); err != nil {
		switch {
		case errors.Is(err, errors2.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "file or knowledge base not found"})
		case errors.Is(err, errors2.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "resource not owned by user"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted successfully"})
}

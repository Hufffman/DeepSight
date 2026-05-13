package middleware

import (
	"net/http"

	"DeepSight/internal/service"

	"github.com/gin-gonic/gin"
)

const ContextKeyUserID = "userID"

func JWTAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := service.ExtractToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
			c.Abort()
			return
		}

		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0
	}

	id, ok := userID.(uint)
	if !ok {
		return 0
	}

	return id
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gin/internal/user/repository"
	"gin/pkg/utils"
)

type Middleware struct {
	repo repository.UserRepository
}

func New(repo repository.UserRepository) *Middleware {
	return &Middleware{repo: repo}
}

func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not have token"})
			c.Abort()
			return
		}

		userId, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not have token"})
			c.Abort()
			return
		}

		user, err := m.repo.FindByID(userId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not have token"})
			c.Abort()
			return
		}

		c.Set("auth_user_id", user.ID)
		c.Next()
	}
}

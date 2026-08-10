package middleware

import (
	"gin/repository"
	"gin/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(repo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

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

		user, err := repo.FindByID(userId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not have token"})
			c.Abort()
			return
		}

		c.Set("auth_user_id", user.ID)
		c.Next()
	}
}

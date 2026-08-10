package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gin/repository"
)

type UserHandler struct {
	repo repository.UserRepository
}

func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.repo.FindByID(uint(id), "Posts")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not found any user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetAuthUserPosts(c *gin.Context) {
	authUserID, exists := c.Get("auth_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userId, ok := authUserID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	user, err := h.repo.FindByID(userId, "Posts")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user.Posts)
}

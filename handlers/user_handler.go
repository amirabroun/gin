package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gin-quickstart/repository"
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

	user, err := h.repo.FindByID(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not found any user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

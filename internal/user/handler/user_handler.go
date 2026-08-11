package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gin/internal/user/service"
	"gin/pkg/utils"
)

type UserHandler struct {
	svc *service.UserService
}

func New(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type LoginRequest struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.svc.GetByID(uint(id), "Posts")

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

	user, err := h.svc.GetByID(userId, "Posts")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user.Posts)
}

func (h *UserHandler) Login(c *gin.Context) {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.Login(request.Mobile, request.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

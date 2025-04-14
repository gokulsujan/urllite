package handler

import (
	"net/http"
	"urllite/service"
	"urllite/types"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	CreateUser(c *gin.Context)
	// GetUserByID(c *gin.Context)
	// GetUserByEmail(c *gin.Context)
	// GetUsers(c *gin.Context)
	// UpdateUserByID(c *gin.Context)
	// DeleteUserByID(c *gin.Context)
}

type userHandler struct {
	userService service.UserService
}

func NewUserHandler() UserHandler {
	userService := service.NewUserService()
	return &userHandler{userService: userService}
}

func (h *userHandler) CreateUser(c *gin.Context) {
	var user types.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid input"})
		return
	}

	if user.Name == "" || user.Email == "" || user.Mobile == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Name, Email, and Mobile are required"})
		return
	}

	if err := h.userService.Create(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to create user", "result": gin.H{"error": err.Error()}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "User created successfully", "result": user})
}

package handler

import (
	"net/http"
	"strings"
	"urllite/auth"
	"urllite/service"
	"urllite/types"
	"urllite/types/dtos"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	CreateUser(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetUsers(c *gin.Context)
	UpdateUserByID(c *gin.Context)
	DeleteUserByID(c *gin.Context)
	ChangePassword(c *gin.Context)

	Signup(c *gin.Context)
	Login(c *gin.Context)
	SendEmailVerificationOtp(c *gin.Context)
	VerifyEmail(c *gin.Context)
	MakeAdmin(c *gin.Context)
}

type userHandler struct {
	userService     service.UserService
	passwordService service.PasswordService
}

func NewUserHandler() UserHandler {
	userService := service.NewUserService()
	passwordService := service.NewPasswordService()
	return &userHandler{userService: userService, passwordService: passwordService}
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

	if appErr := h.userService.Create(&user); appErr != nil {
		appErr.HttpResponse(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "User created successfully", "result": user})
}

func (h *userHandler) GetUsers(c *gin.Context) {
	filter := types.UserFilter{
		Name:   strings.TrimSpace(c.Query("name")),
		Mobile: strings.TrimSpace(c.Query("mobile")),
		Email:  strings.TrimSpace(c.Query("email")),
		Status: strings.TrimSpace(c.Query("status")),
	}
	users, appErr := h.userService.GetUsers(filter)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Succesfully get the users", "result": gin.H{"users": users}})
}

func (h *userHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, appErr := h.userService.GetUserByID(id)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "result": user})
}

func (h *userHandler) UpdateUserByID(c *gin.Context) {
	id := c.Param("id")

	var user types.User
	err := c.ShouldBindJSON(&user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable read the request body", "result": gin.H{"error": err.Error()}})
		return
	}

	appErr := h.userService.UpdateUserByID(id, user)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "User updated successfully!"})
}

func (h *userHandler) DeleteUserByID(c *gin.Context) {
	id := c.Param("id")
	appErr := h.userService.DeleteUserByID(id)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "User deleted successfully!"})
}

func (h *userHandler) Signup(c *gin.Context) {
	var signupReq dtos.SignupDTO
	if err := c.ShouldBindJSON(&signupReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request", "result": gin.H{"error": map[string]interface{}{"errot": err.Error()}}})
		return
	}

	if signupReq.ConfirmPassword != signupReq.Password {
		c.JSON(http.StatusNotAcceptable, gin.H{"status": "failed", "message": "Password and confirm passwords are not same"})
		return
	}

	user := &types.User{Name: signupReq.Name, Email: signupReq.Email, Mobile: signupReq.Mobile, Status: "registered"}
	appErr := h.userService.Create(user)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	_, appErr = h.passwordService.Create(signupReq.Password, user.ID.String())
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "Signup successfull!! Please verify the email."})
}

func (h *userHandler) Login(c *gin.Context) {
	var loginReq dtos.LoginDTO
	err := c.ShouldBindJSON(&loginReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request", "result": gin.H{"error": err.Error()}})
		return
	}

	user, appErr := h.userService.GetUserByEmail(loginReq.Email)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	password, appErr := h.passwordService.GetPasswordByUserID(user.ID.String())
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}
	isPasswordValid := h.passwordService.VerifyPassword(loginReq.Password, password)

	if isPasswordValid {
		accessToken, appErr := h.userService.GenerateUserAccessToken(user, c.Request.Context())
		if appErr != nil {
			appErr.HttpResponse(c)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Token generated", "access_token": accessToken})
	} else {
		c.JSON(http.StatusNotAcceptable, gin.H{"status": "failed", "message": "Incorrect Password"})
	}
}

func (h *userHandler) ChangePassword(c *gin.Context) {
	// TODO -> Get User details from the token
	current_user := auth.CurrentUserFromContext(c)

	var changePasswordRequest dtos.PasswordChangeDTO
	err := c.ShouldBindJSON(&changePasswordRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request", "result": gin.H{"error": err.Error()}})
		return
	}

	appErr := h.passwordService.ChangePassword(current_user.Email, changePasswordRequest.Password, changePasswordRequest.NewPassword)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "Password changed successfully"})
}

func (h *userHandler) SendEmailVerificationOtp(c *gin.Context) {
	var emailVerificationDto dtos.EmailVerificationDTO
	err := c.ShouldBindJSON(&emailVerificationDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request", "result": gin.H{"error": err.Error()}})
		return
	}

	user, appErr := h.userService.GetUserByEmail(emailVerificationDto.Email)
	if err != nil {
		appErr.HttpResponse(c)
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "No user found"})
		return
	}

	appErr = h.userService.SendEmailVerificationOtp(emailVerificationDto.Email)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "OTP Delivered successfully"})
}

func (h *userHandler) VerifyEmail(c *gin.Context) {
	var emailVerificationDto dtos.EmailVerificationDTO
	err := c.ShouldBindJSON(&emailVerificationDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request", "result": gin.H{"error": err.Error()}})
		return
	}

	user, appErr := h.userService.GetUserByEmail(emailVerificationDto.Email)
	if err != nil {
		appErr.HttpResponse(c)
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "No user found"})
		return
	}

	appErr = h.userService.VerifyEmail(emailVerificationDto.Email, emailVerificationDto.Otp)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Email veified successfully"})
}

func (h *userHandler) MakeAdmin(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "No user id found"})
		return
	}

	appErr := h.userService.MakeAdmin(userId)
	if appErr != nil {
		appErr.HttpResponse(c)
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "User role changed to admin"})
}

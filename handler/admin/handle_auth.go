package admin_handlers

import (
	"net/http"
	"urllite/service"
	"urllite/types/dtos"
	"urllite/utils"

	"github.com/gin-gonic/gin"
)

type AdminAuthHandler interface {
	Login(c *gin.Context)
	Dashboard(c *gin.Context)
}

type adminAuthHandler struct {
	userService     service.UserService
	passwordService service.PasswordService
	adminService    service.AdminService
}

func NewAdminAuthHandler() AdminAuthHandler {
	userService := service.NewUserService()
	passwordService := service.NewPasswordService()
	adminService := service.NewAdminService()
	return adminAuthHandler{userService: userService, passwordService: passwordService, adminService: adminService}
}

func (h adminAuthHandler) Login(c *gin.Context) {
	var loginDto dtos.LoginDTO
	err := c.ShouldBindBodyWithJSON(&loginDto)
	if err != nil {
		utils.HttpResponse(c, http.StatusBadRequest, "Inavlid request", gin.H{"error": err.Error()})
		return
	}

	user, appErr := h.userService.GetUserByEmail(loginDto.Email)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	if user.Role != "admin" {
		utils.HttpResponse(c, http.StatusUnauthorized, "Invalid Credentials", nil)
		return
	}

	password, appErr := h.passwordService.GetPasswordByUserID(user.ID.String())
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	if h.passwordService.VerifyPassword(loginDto.Password, password) {
		token, appErr := h.userService.GenerateUserAccessToken(user, c.Request.Context())
		if appErr != nil {
			appErr.HttpResponse(c)
			return
		}

		utils.HttpResponse(c, http.StatusAccepted, "Login successfull", gin.H{"access_token": token})
		return
	}

	utils.HttpResponse(c, http.StatusUnauthorized, "Invalid Credentials", nil)
}

func (h adminAuthHandler) Dashboard(c *gin.Context) {
	dashboard, appErr := h.adminService.Dashboard()
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	utils.HttpResponse(c, http.StatusOK, "Dashboard data fetched", gin.H{"dashboard": dashboard})
}

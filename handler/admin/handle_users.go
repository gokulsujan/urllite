package admin_handlers

import (
	"net/http"
	"urllite/service"
	"urllite/utils"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler interface {
	UserDashboardStats(c *gin.Context)
}

type adminUserHandler struct {
	userService  service.UserService
	adminService service.AdminService
}

func NewAdminUserHandler() AdminUserHandler {
	userSvc := service.NewUserService()
	adminSvc := service.NewAdminService()
	return adminUserHandler{userService: userSvc, adminService: adminSvc}
}

func (h adminUserHandler) UserDashboardStats(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.HttpResponse(c, http.StatusBadRequest, "User Id is empty", nil)
	}
	stats, appErr := h.userService.UserDashboardStats(id)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	utils.HttpResponse(c, http.StatusOK, "User stats fetched", map[string]interface{}{"user_stats": stats})
}

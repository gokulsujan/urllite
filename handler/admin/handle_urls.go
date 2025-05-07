package admin_handlers

import (
	"net/http"
	"urllite/service"
	"urllite/utils"

	"github.com/gin-gonic/gin"
)

type AdminUrlHandler interface {
	UrlsByUserID(c *gin.Context)
}

type adminUrlHandler struct {
	urlService   service.UrlService
	adminService service.AdminService
}

func NewAdminUrlHandler() AdminUrlHandler {
	urlSvc := service.NewUrlService()
	adminSvc := service.NewAdminService()
	return adminUrlHandler{urlService: urlSvc, adminService: adminSvc}
}

func (h adminUrlHandler) UrlsByUserID(c *gin.Context) {
	id := c.Param("id")
	urls, appErr := h.urlService.GetUrlsOfUser(id)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	utils.HttpResponse(c, http.StatusOK, "User url fetched", map[string]interface{}{"urls": urls})
}

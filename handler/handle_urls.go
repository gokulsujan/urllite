package handler

import (
	"net/http"
	"urllite/service"
	"urllite/types/dtos"

	"github.com/gin-gonic/gin"
)

type UrlHandler interface {
	Create(c *gin.Context)
	RedirectToLongUrl(c *gin.Context)
	GetUrlByID(c *gin.Context)
	GetURLs(c *gin.Context)
	DeleteURLById(c *gin.Context)
}
type urlHandler struct {
	urlService    service.UrlService
	urlLogService service.UrlLogService
}

func NewUrlHandler() UrlHandler {
	urlService := service.NewUrlService()
	logService := service.NewUrlLogService()
	return &urlHandler{urlService: urlService, urlLogService: logService}
}

func (u *urlHandler) Create(c *gin.Context) {
	var urlDto dtos.UrlDTO
	err := c.ShouldBindJSON(&urlDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request body", "result": gin.H{"error": err.Error()}})
		return
	}

	curent_user, ok := c.Get("current_user_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "No user data found in the context"})
	}
	url, appErr := u.urlService.CreateUrl(urlDto.LongUrl, curent_user.(string))
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "Url created successfully", "result": gin.H{"url": url}})
}

func (u *urlHandler) RedirectToLongUrl(c *gin.Context) {
	shortUrl := c.Param("short_url")
	url, appErr := u.urlService.GetUrlByShortUrl(shortUrl)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	if url == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "No url found"})
		return
	}

	u.urlLogService.CreateUrlLogByUrl(url, c.ClientIP())
	c.Redirect(http.StatusFound, url.LongUrl)
}

func (u *urlHandler) GetUrlByID(c *gin.Context) {
	urlId := c.Param("id")
	current_user_id, ok := c.Get("current_user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to get current user id from context"})
		return
	}
	url, appErr := u.urlService.GetUrlByID(urlId, current_user_id.(string))
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	currentUserID, ok := c.Get("current_user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to get current user id from context"})
		return
	}

	if url.UserID.String() != currentUserID.(string) {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "No url found"})
		return
	}

	logs, appErr := u.urlService.GetUrlLogsByUrl(url)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Url fetched successfully", "result": gin.H{"url": url, "logs": logs}})
}

func (u *urlHandler) GetURLs(c *gin.Context) {
	user_id, ok := c.Get("current_user_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "No user data found in the context"})
	}
	urls, appErr := u.urlService.GetUrlsOfUser(user_id.(string))
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Urls fetched successfully", "result": gin.H{"urls": urls}})
}

func (u *urlHandler) DeleteURLById(c *gin.Context) {
	urlId := c.Param("id")
	current_user_id, ok := c.Get("current_user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to get current user id from context"})
		return
	}
	appErr := u.urlLogService.DeleteUrlLogByUrl(urlId)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}
	appErr = u.urlService.DeleteUrlById(urlId, current_user_id.(string))
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Url deleted successfully"})
}

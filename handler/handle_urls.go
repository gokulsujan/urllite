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
	urlService service.UrlService
}

func NewUrlHandler() UrlHandler {
	urlService := service.NewUrlService()
	return &urlHandler{urlService: urlService}
}

func (u *urlHandler) Create(c *gin.Context) {
	var urlDto dtos.UrlDTO
	err := c.ShouldBindJSON(&urlDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request body", "result": gin.H{"error": err.Error()}})
		return
	}

	url, appErr := u.urlService.CreateUrl(urlDto.LongUrl)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "success", "message": "Url created successfully", "result": gin.H{"url": url}})
}

func (u *urlHandler) RedirectToLongUrl(c *gin.Context) {
	var urlDto dtos.UrlDTO
	err := c.ShouldBindJSON(&urlDto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request body", "result": gin.H{"error": err.Error()}})
		return
	}

	url, appErr := u.urlService.GetUrlByShortUrl(urlDto.ShortUrl)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.Redirect(http.StatusPermanentRedirect, url.LongUrl)
}

func (u *urlHandler) GetUrlByID(c *gin.Context) {
	urlId := c.Param("id")
	url, appErr := u.urlService.GetUrlByID(urlId)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Url fetched successfully", "result": gin.H{"url": url}})
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
	appErr := u.urlService.DeleteUrlById(urlId)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Url deleted successfully"})
}

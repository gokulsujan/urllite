package handler

import (
	"net/http"
	"strings"
	"urllite/service"
	"urllite/types/dtos"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

type UrlHandler interface {
	Create(c *gin.Context)
	RedirectToLongUrl(c *gin.Context)
	GetUrlByID(c *gin.Context)
	GetURLs(c *gin.Context)
	DeleteURLById(c *gin.Context)
	GetUrlLogsByUrl(c *gin.Context)
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

	metadata := fetchMetadata(url.LongUrl)
	c.HTML(http.StatusOK, "preview.html", gin.H{
		"Title":       metadata.Title,
		"Description": metadata.Description,
		"Image":       metadata.Image,
		"Url":         url.LongUrl,
	})
}

func (u *urlHandler) GetUrlByID(c *gin.Context) {
	urlId := c.Param("id")
	current_user_id, ok := c.Get("current_user_id")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to get current user id from context"})
		return
	}
	url, appErr := u.urlService.GetUrlByID(urlId)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	currentUserRole, ok := c.Get("current_user_role")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Unable to get current role from context"})
		return
	}

	if currentUserRole.(string) != "admin" && (url.UserID.String() != current_user_id.(string)) {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "No url found"})
		return
	}

	urlMetadata, appErr := u.urlService.GetUrlDatas(url)
	if appErr != nil {
		appErr.HttpResponse(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Url fetched successfully", "result": gin.H{"url": url, "meta": urlMetadata}})
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

func (u *urlHandler) GetUrlLogsByUrl(c *gin.Context) {
	urlId := c.Param("id")
	url, appErr := u.urlService.GetUrlByID(urlId)
	if appErr != nil {
		appErr.HttpResponse(c)
	}

	logs, appErr := u.urlService.GetUrlLogsByUrl(url)
	if appErr != nil {
		appErr.HttpResponse(c)
	}

	responseMessage := "Logs successfully fetched"
	if logs == nil {
		responseMessage = "No logs available"
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": responseMessage, "result": gin.H{"logs": logs}})

}

func fetchMetadata(targetUrl string) *dtos.UrlMetadata {
	resp, err := http.Get(targetUrl)
	if err != nil || resp.StatusCode != http.StatusOK {
		return &dtos.UrlMetadata{}
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	metadata := &dtos.UrlMetadata{}

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()
		if t.Type == html.StartTagToken && t.Data == "meta" {
			var prop, content string
			for _, attr := range t.Attr {
				switch strings.ToLower(attr.Key) {
				case "property", "name":
					prop = attr.Val
				case "content":
					content = attr.Val
				}
			}

			switch prop {
			case "og:title":
				metadata.Title = content
			case "og:description":
				metadata.Description = content
			case "og:image":
				metadata.Image = content
			}
		}
	}

	return metadata
}

package types

import "github.com/gin-gonic/gin"

type ApplicationError struct {
	Message        string `json:"message"`
	HttpStatusCode int    `json:"code"`
	Err          error  `json:"error"`
}

func (e *ApplicationError) HttpResponse(c *gin.Context) {
	if e.Err == nil {
		c.JSON(e.HttpStatusCode, gin.H{
			"status":  "failed",
			"message": e.Message,
		})
		return
	}

	c.JSON(e.HttpStatusCode, gin.H{
		"status":  "failed",
		"message": e.Message,
		"result":  gin.H{"error": e.Err.Error()},
	})
}

func (e *ApplicationError) Error() error {
	return e.Err
}

package response

import "github.com/gin-gonic/gin"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error": ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

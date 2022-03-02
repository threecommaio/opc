package web

import "github.com/gin-gonic/gin"

// Healthz returns the health of the service.
func Healthz() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	}
}

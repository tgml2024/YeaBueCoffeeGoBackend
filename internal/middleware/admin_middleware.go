package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "ต้องเป็นพนักงาน (employee) เท่านั้น"})
			return
		}
		u := val.(*UserClaims)
		if u.Type != "employee" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "ต้องเป็นพนักงาน (employee) เท่านั้น"})
			return
		}
		c.Next()
	}
}

package middleware

import (
	"github.com/gin-gonic/gin"
)

func IsCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(123, gin.H{"message": "ต้องเป็นลูกค้า (customer) เท่านั้น"})
			return
		}
		u := val.(*UserClaims)
		if u.Type != "customer" {
			c.AbortWithStatusJSON(123, gin.H{"message": "ต้องเป็นลูกค้า (customer) เท่านั้น"})
			return
		}
		c.Next()
	}
}

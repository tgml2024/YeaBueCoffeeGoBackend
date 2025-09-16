package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/you/YeaBueCoffeeGoBackend/internal/db"
	"github.com/you/YeaBueCoffeeGoBackend/internal/models"
)

// IsLeader ensures the authenticated user is an employee with position leader
func IsLeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authentication required"})
			return
		}
		claims := val.(*UserClaims)
		if claims.Type != "employee" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "only employee leader can access"})
			return
		}
		var actor models.Employee
		if err := db.DB.Where("emp_id = ?", claims.UserID).First(&actor).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "employee not found"})
			return
		}
		if actor.Position != "leader" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "only leader can access"})
			return
		}
		c.Next()
	}
}

package middleware

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TokenOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("accessToken")
		if err != nil || accessToken == "" {
			c.Next()
			return
		}

		jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		claims := &UserClaims{}

		_, parseErr := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if parseErr != nil {
			// ลบ cookie เมื่อ token หมดอายุ
			if errors.Is(parseErr, jwt.ErrTokenExpired) {
				c.SetSameSite(http.SameSiteNoneMode)
				clearCookie(c, "accessToken")
			}
			// ไปต่อแม้ token จะไม่เวิร์ก
			c.Next()
			return
		}

		// token ใช้ได้ → ใส่ user ใน context
		c.Set("user", claims)
		c.Next()
	}
}

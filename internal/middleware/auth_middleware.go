package middleware

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/you/YeaBueCoffeeGoBackend/internal/db"
	"github.com/you/YeaBueCoffeeGoBackend/internal/models"
)

type UserClaims struct {
	UserID string `json:"userId"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func AuthenticateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, errA := c.Cookie("accessToken")
		refreshToken, errR := c.Cookie("refreshToken")

		if errA != nil && errR != nil {
			log.Println("Both access and refresh tokens are missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No tokens provided"})
			return
		}

		jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		refreshSecret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))

		// ถ้ามี access token ลอง verify ก่อน
		if errA == nil && accessToken != "" {
			claims := &UserClaims{}
			_, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err == nil {
				// ok
				c.Set("user", claims)
				// update last_access if sessionId exists
				if sid, e := c.Cookie("sessionId"); e == nil && sid != "" {
					_ = db.DB.Model(&models.Authen{}).Where("session_id = ?", sid).Update("last_access", time.Now()).Error
				}
				c.Next()
				return
			}
			// access token invalid → ล้าง cookie ทั้งคู่ ตามพฤติกรรมเดิม
			clearCookie(c, "accessToken")
			clearCookie(c, "refreshToken")
		}

		// ใช้ refresh token ออก access ใหม่
		if errR == nil && refreshToken != "" {
			claims := &UserClaims{}
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
				return refreshSecret, nil
			})
			if err != nil {
				log.Println("Refresh token verification failed:", err)
				clearCookie(c, "accessToken")
				clearCookie(c, "refreshToken")
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid refresh token"})
				return
			}

			// ออก access token ใหม่ 1 ชั่วโมง
			newAccess := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
				UserID: claims.UserID,
				Type:   claims.Type,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			})
			signed, err := newAccess.SignedString(jwtSecret)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cannot sign access token"})
				return
			}

			setCookie(c, "accessToken", signed, time.Hour)
			// ตั้ง user ใน context แล้วไปต่อ
			c.Set("user", claims)
			c.Next()
			return
		}

		// มาถึงตรงนี้ = ไม่มี access และใช้ refresh ไม่ได้
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Access token missing, refresh required"})
	}
}

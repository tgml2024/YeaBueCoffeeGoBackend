package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func setCookie(c *gin.Context, name, value string, maxAge time.Duration) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	domain := os.Getenv("COOKIE_DOMAIN")

	sameSite := "Strict"
	switch os.Getenv("COOKIE_SAMESITE") {
	case "lax":
		sameSite = "Lax"
	case "none":
		sameSite = "None"
	}

	cookie := fmt.Sprintf("%s=%s; Path=/; Max-Age=%d; Domain=%s; SameSite=%s",
		name, value, int(maxAge.Seconds()), domain, sameSite)

	if secure {
		cookie += "; Secure"
	}
	cookie += "; HttpOnly"

	c.Writer.Header().Add("Set-Cookie", cookie)
}

func clearCookie(c *gin.Context, name string) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	domain := os.Getenv("COOKIE_DOMAIN")

	c.SetCookie(name, "", -1, "/", domain, secure, true)
}

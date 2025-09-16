package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/you/YeaBueCoffeeGoBackend/internal/db"
	"github.com/you/YeaBueCoffeeGoBackend/internal/middleware"
	"github.com/you/YeaBueCoffeeGoBackend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Type     string `json:"type" binding:"required"`
}

type UserClaims struct {
	UserID string `json:"userId"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
		return
	}

	switch req.Type {
	case "employee":
		var emp models.Employee
		if err := db.DB.Where("username = ?", req.Username).First(&emp).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "username or password wrong"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(emp.Password), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "username or password wrong"})
			return
		}
		accessToken, refreshToken, err := issueTokens(emp.EmpID, "employee")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "cannot issue token"})
			return
		}

		// create session row
		sessionID := generateSessionID()
		now := time.Now()
		_ = db.DB.Create(&models.Authen{
			SessionID:  sessionID,
			StartDate:  now,
			LastAccess: now,
			UserID:     emp.EmpID,
		}).Error

		setAuthCookies(c, accessToken, refreshToken)
		// set session cookie (30d)
		secure := os.Getenv("COOKIE_SECURE") == "true"
		domain := os.Getenv("COOKIE_DOMAIN")
		c.SetCookie("sessionId", sessionID, 30*24*3600, "/", domain, secure, true)
		c.JSON(http.StatusOK, gin.H{"message": "login success", "type": "employee"})
		return
	case "customer":
		var cust models.Customer
		if err := db.DB.Where("username = ?", req.Username).First(&cust).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "username or password wrong"})
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(cust.Password), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "username or password wrong"})
			return
		}
		accessToken, refreshToken, err := issueTokens(cust.CustID, "customer")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "cannot issue token"})
			return
		}

		// create session row
		sessionID := generateSessionID()
		now := time.Now()
		_ = db.DB.Create(&models.Authen{
			SessionID:  sessionID,
			StartDate:  now,
			LastAccess: now,
			UserID:     cust.CustID,
		}).Error

		setAuthCookies(c, accessToken, refreshToken)
		// set session cookie (30d)
		secure := os.Getenv("COOKIE_SECURE") == "true"
		domain := os.Getenv("COOKIE_DOMAIN")
		c.SetCookie("sessionId", sessionID, 30*24*3600, "/", domain, secure, true)
		c.JSON(http.StatusOK, gin.H{"message": "login success", "type": "customer"})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid type, must be 'customer' or 'employee'"})
		return
	}
}

func Logout(c *gin.Context) {
	// mark session end
	if sid, err := c.Cookie("sessionId"); err == nil && sid != "" {
		_ = db.DB.Model(&models.Authen{}).Where("session_id = ?", sid).Update("end_date", time.Now()).Error
	}
	clearCookie(c, "accessToken")
	clearCookie(c, "refreshToken")
	secure := os.Getenv("COOKIE_SECURE") == "true"
	domain := os.Getenv("COOKIE_DOMAIN")
	c.SetCookie("sessionId", "", -1, "/", domain, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func issueTokens(uid string, userType string) (string, string, error) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	refreshSecret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		UserID: uid,
		Type:   userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		UserID: uid,
		Type:   userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	a, err := access.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	r, err := refresh.SignedString(refreshSecret)
	if err != nil {
		return "", "", err
	}

	return a, r, nil
}

func setAuthCookies(c *gin.Context, access, refresh string) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	domain := os.Getenv("COOKIE_DOMAIN")

	c.SetCookie("accessToken", access, 3600, "/", domain, secure, true)
	c.SetCookie("refreshToken", refresh, 30*24*3600, "/", domain, secure, true)
}

func clearCookie(c *gin.Context, name string) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	domain := os.Getenv("COOKIE_DOMAIN")
	c.SetCookie(name, "", -1, "/", domain, secure, true)
}

func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(b)
}

// Registration

type registerCustomerReq struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type registerEmployeeReq struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Position  string `json:"position"`
}

func RegisterCustomer(c *gin.Context) {
	var req registerCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
		return
	}
	var exists int64
	db.DB.Model(&models.Customer{}).Where("username = ?", req.Username).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "username already exists"})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "cannot hash password"})
		return
	}
	id := ("C" + generateSessionID())[:10]
	now := time.Now()
	cust := models.Customer{
		CustID:    id,
		Username:  req.Username,
		Password:  string(hashed),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		AddDate:   now,
		AddUser:   "system",
	}
	if err := db.DB.Create(&cust).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "cannot create customer"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "customer registered", "cust_id": cust.CustID})
}

func RegisterEmployee(c *gin.Context) {
	// require authenticated leader (employee)
	val, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "authentication required"})
		return
	}
	claims := val.(*middleware.UserClaims)
	if claims.Type != "employee" {
		c.JSON(http.StatusForbidden, gin.H{"message": "only employee leader can register employees"})
		return
	}
	var actor models.Employee
	if err := db.DB.Where("emp_id = ?", claims.UserID).First(&actor).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "employee not found"})
		return
	}
	if actor.Position != "leader" {
		c.JSON(http.StatusForbidden, gin.H{"message": "only leader can register employees"})
		return
	}

	var req registerEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
		return
	}
	var exists2 int64
	db.DB.Model(&models.Employee{}).Where("username = ?", req.Username).Count(&exists2)
	if exists2 > 0 {
		c.JSON(http.StatusConflict, gin.H{"message": "username already exists"})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "cannot hash password"})
		return
	}
	if req.Position == "" {
		req.Position = "employee"
	}
	// allow only "employee" or "leader"
	switch req.Position {
	case "employee", "leader":
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid position, allowed: employee | leader"})
		return
	}
	id := ("E" + generateSessionID())[:10]
	now := time.Now()
	emp := models.Employee{
		EmpID:     id,
		Username:  req.Username,
		Password:  string(hashed),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Position:  req.Position,
		AddDate:   now,
		AddUser:   actor.EmpID,
	}
	if err := db.DB.Create(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "cannot create employee"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "employee registered", "emp_id": emp.EmpID})
}

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/you/YeaBueCoffeeGoBackend/internal/controllers"
	"github.com/you/YeaBueCoffeeGoBackend/internal/middleware"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// ใส่ token optional กลาง ๆ ถ้าอยากอ่าน user จาก cookie (เช่นแสดงหน้าโปรไฟล์ public)
	r.Use(middleware.TokenOptional())

	api := r.Group("/api")

	// auth
	api.POST("/login", controllers.Login)
	api.POST("/logout", controllers.Logout)
	api.POST("/register/customer", controllers.RegisterCustomer)

	// employee registration must be authenticated and leader-only
	api.POST("/register/employee", middleware.AuthenticateToken(), middleware.IsEmployee(), middleware.IsLeader(), controllers.RegisterEmployee)

	// เส้นทางที่ต้องตรวจ token แบบบังคับ (เหมือน authenticateToken)
	protected := api.Group("/")
	protected.Use(middleware.AuthenticateToken())

	// customer routes
	customer := protected.Group("/customer")
	customer.Use(middleware.IsCustomer())
	customer.GET("/me", controllers.Me)

	// employee routes
	employee := protected.Group("/employee")
	employee.Use(middleware.IsEmployee())
	employee.GET("/me", controllers.Me)

	return r
}

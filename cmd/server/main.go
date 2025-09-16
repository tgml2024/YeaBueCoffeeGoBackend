package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/you/YeaBueCoffeeGoBackend/internal/db"
	"github.com/you/YeaBueCoffeeGoBackend/internal/routes"
)

func main() {
	_ = godotenv.Load() // โหลด .env ถ้ามี

	// เปิด DB
	if err := db.Connect(); err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	// ทำ auto-migrate model
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("DB migrate error: %v", err)
	}

	// สร้าง router + ผูก route
	r := routes.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s ...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

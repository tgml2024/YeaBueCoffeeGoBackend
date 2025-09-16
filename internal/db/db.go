package db

import (
	"github.com/you/YeaBueCoffeeGoBackend/internal/config"
	"github.com/you/YeaBueCoffeeGoBackend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	cfg := config.Load()
	dsn := config.DSN(cfg)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return err
}

func AutoMigrate() error {
	return DB.AutoMigrate(&models.Customer{}, &models.Employee{}, &models.Authen{})
}

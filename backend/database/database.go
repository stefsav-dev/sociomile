package database

import (
	"backend/model"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectionDB() {
	dsn := "root@tcp(127.0.0.1:3306)/sociomile_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	defer db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Message{}, &model.BlacklistedToken{})
	fmt.Println("Database terkoneksi & migrasi berhasil!")
	DB = db
}

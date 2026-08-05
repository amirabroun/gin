package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/joho/godotenv"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID               uint `gorm:"primaryKey"`
	Contact          string
	SubscriberId     uint
	SubscriberUserId uint
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) getUser(ctx *gin.Context) {
	user := User{}

	h.db.Where("id = ?", ctx.Param("id")).First(&user)

	ctx.JSON(http.StatusOK, user)
}

func initDB() *gorm.DB {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err error
	var db *gorm.DB

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	return db
}

func main() {
	db := initDB()

	h := NewHandler(db)

	router := gin.Default()

	router.SetTrustedProxies(nil)

	router.GET("users/:id", h.getUser)

	router.Run(":8090")
}

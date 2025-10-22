package main

import (
	"bobri/config"
	"bobri/helpers"
	"bobri/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

func main() {

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("error while connecting to db:%v", err)
	}
	defer db.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // именно как в твоем сообщении
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	secret := os.Getenv("JWT_SECRET")
	jwtMaker := helpers.NewJWTMaker([]byte(secret), 24*time.Hour)

	routes.AuthRoutes(r, db, jwtMaker)

	go func() {
		log.Println("🚀 Сервер запущен на порту 8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("❌ Ошибка запуска сервера: %v", err)
		}
	}()

}

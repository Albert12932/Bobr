package main

import (
	"bobri/config"
	"bobri/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	routes.AuthRoutes(r, db)

	go func() {
		log.Println("🚀 Сервер запущен на порту 8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("❌ Ошибка запуска сервера: %v", err)
		}
	}()

	// Канал для получения сигналов ОС
	quit := make(chan os.Signal, 1)
	// Регистрируем интересующие сигналы
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем основную горутину, пока не получим сигнал
	<-quit
	log.Println("Shutting down server...")

}

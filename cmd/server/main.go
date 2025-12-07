package main

import (
	_ "bobri/docs"
	"bobri/internal/api/routes"
	"bobri/internal/config"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

func main() {

	// Подключаемся к БД
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Error while connecting to db:%v", err)
	}
	defer db.Close()

	// Получаем secret из .env и создаем JWTMaker
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	AccessJwtMaker := helpers.NewJWTMaker([]byte(secret), 15*time.Minute)

	fromEmail := os.Getenv("FROM_EMAIL")
	emailPass := os.Getenv("EMAIL_PASSWORD")

	if fromEmail == "" || emailPass == "" {
		log.Fatal("Email credentials are missing in ENV")
	}

	// Создаем движок gin для работы с HTTP и регистрируем роутеры
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	routes.AuthRoutes(engine, db, AccessJwtMaker, models.EmailAuth{
		EmailFrom: fromEmail,
		EmailPass: emailPass})
	routes.AdminRoutes(engine, db, AccessJwtMaker)
	routes.UserRoutes(engine, db, AccessJwtMaker)

	// Запускаем движок
	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("Error while starting server: %v", err)
	}

}

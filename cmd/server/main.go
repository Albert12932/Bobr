package main

import (
	_ "bobri/docs"
	"bobri/internal/api/routes"
	"bobri/internal/config"
	"bobri/pkg/helpers"
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

	// Создаем движок gin для работы с HTTP и регистрируем роутеры
	engine := gin.Default()

	routes.AuthRoutes(engine, db, AccessJwtMaker)
	routes.HelperRoutes(engine, db, AccessJwtMaker)

	// Запускаем движок
	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("Error while starting server: %v", err)
	}

}

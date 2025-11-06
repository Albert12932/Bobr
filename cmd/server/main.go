package main

import (
	_ "bobri/docs"
	"bobri/pkg/api/routes"
	"bobri/pkg/config"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

func main() {

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Error while connecting to db:%v", err)
	}
	defer db.Close()

	engine := gin.Default()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	AccessJwtMaker := helpers.NewJWTMaker([]byte(secret), 30*time.Minute)

	routes.AuthRoutes(engine, db, AccessJwtMaker)

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("Error while starting server: %v", err)
	}

}

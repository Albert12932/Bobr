package routes

import (
	"bobri/controllers"
	"bobri/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool, jwtMaker *helpers.JWTMaker) {
	r.POST("/auth/check", controllers.AuthCheck(db))
	r.POST("/auth/:token", controllers.RegisterByToken(db, jwtMaker))
}

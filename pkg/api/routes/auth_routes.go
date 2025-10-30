package routes

import (
	"bobri/pkg/api/controllers"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool, jwtMaker *helpers.JWTMaker) {
	r.POST("/auth/check", controllers.AuthCheck(db))
	r.POST("/auth/:token", controllers.RegisterByToken(db, jwtMaker))
	r.POST("/auth/login", controllers.Login(db, jwtMaker))
	r.DELETE("/helper/deleteUser", controllers.DeleteUser(db))
	r.GET("/helper/students", controllers.GetStudents(db))
	r.GET("/helper/users", controllers.GetUsers(db))
}

package routes

import (
	"bobri/pkg/api/controllers"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool, AccessJwtMaker *helpers.JWTMaker) {
	r.POST("/auth/check", controllers.AuthCheck(db))
	r.POST("/auth/register", controllers.RegisterByToken(db))
	r.POST("/auth/login", controllers.Login(db, AccessJwtMaker))
	r.DELETE("/helper/deleteUser", controllers.DeleteUser(db))
	r.GET("/helper/students", controllers.GetStudents(db))
	r.GET("/helper/users", controllers.GetUsers(db))
	r.POST("/auth/refresh", controllers.RefreshToken(db, AccessJwtMaker))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.POST("auth/reset_password", controllers.ResetPassword(db))
	r.POST("auth/set_new_password", controllers.SetNewPassword(db))
}

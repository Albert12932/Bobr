package routes

import (
	"bobri/internal/api/controllers"
	"bobri/internal/middleware"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool, accessJwtMaker *helpers.JWTMaker) {

	// Хэндлеры аутентификации
	r.POST("/auth/check", controllers.AuthCheck(db))
	r.POST("/auth/register", controllers.RegisterByToken(db))
	r.POST("/auth/login", controllers.Login(db, accessJwtMaker))
	r.POST("auth/reset_password", controllers.ResetPassword(db))
	r.POST("auth/set_new_password", controllers.SetNewPassword(db))
	r.POST("/auth/refresh", controllers.RefreshToken(db, accessJwtMaker))

	// Хэндлер swagger-а
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

}

func HelperRoutes(r *gin.Engine, db *pgxpool.Pool, accessJWTMaker *helpers.JWTMaker) {

	// Создаем группу защищенных хэндлеров
	protectedHelper := r.Group("/helper")
	protectedHelper.Use(middleware.AuthenticationMiddleware(accessJWTMaker))
	protectedHelper.DELETE("/deleteUser", controllers.DeleteUser(db))
	protectedHelper.GET("/students", controllers.GetStudents(db))
	protectedHelper.GET("/users", controllers.GetUsers(db))
}

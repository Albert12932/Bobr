package routes

import (
	"bobri/internal/api/controllers/auth"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool, accessJwtMaker *helpers.JWTMaker) {

	// Хэндлеры аутентификации
	r.POST("/auth/check", auth.CheckStudent(db))
	r.POST("/auth/register", auth.RegisterByToken(db, accessJwtMaker))
	r.POST("/auth/login", auth.Login(db, accessJwtMaker))
	r.POST("/auth/reset_password", auth.ResetPassword(db))
	r.POST("/auth/set_new_password", auth.SetNewPassword(db))
	r.POST("/auth/refresh", auth.RefreshToken(db, accessJwtMaker))

	// Хэндлер swagger-а
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

}

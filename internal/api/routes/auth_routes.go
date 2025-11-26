package routes

import (
	"bobri/internal/api/controllers/auth"
	"bobri/internal/api/repositories"
	"bobri/internal/api/services"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool, accessJwtMaker *helpers.JWTMaker, emailAuth models.EmailAuth) {

	// Репозитории и сервисы
	userRepo := repositories.NewUserRepository(db)
	refreshTokensRepo := repositories.NewRefreshTokensRepository(db)
	studentsRepo := repositories.NewStudentsRepository(db)
	resetPasswordRepo := repositories.NewResetPasswordRepository(db)
	tokenProvider := services.NewTokenProvider(accessJwtMaker, refreshTokensRepo)
	emailProvider := services.NewEmailProvider(emailAuth)
	authService := services.NewStudentsService(studentsRepo, db)
	registerService := services.NewRegisterService(userRepo, tokenProvider, studentsRepo, db)
	loginService := services.NewLoginService(tokenProvider, userRepo, db)
	refreshService := services.NewRefreshTokensService(db, refreshTokensRepo, accessJwtMaker)
	resetPasswordService := services.NewResetPasswordService(resetPasswordRepo, emailProvider, db)

	authHandlersGroup := r.Group("/auth")
	// Хэндлеры аутентификации
	authHandlersGroup.POST("/check", auth.CheckStudent(authService))
	authHandlersGroup.POST("/register", auth.RegisterByToken(registerService))
	authHandlersGroup.POST("/login", auth.Login(loginService))
	authHandlersGroup.POST("/reset_password", auth.ResetPassword(resetPasswordService))
	authHandlersGroup.POST("/set_new_password", auth.SetNewPassword(resetPasswordService))
	authHandlersGroup.POST("/refresh", auth.RefreshToken(refreshService))

	// Хэндлер swagger-а
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

}

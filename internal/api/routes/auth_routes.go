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
	// создаем UoW
	uow := repositories.NewUoW(db)

	// репозитории
	userRepo := repositories.NewUserRepository(db)
	refreshTokensRepo := repositories.NewRefreshTokensRepository(db)
	studentsRepo := repositories.NewStudentsRepository(db)
	resetPasswordRepo := repositories.NewResetPasswordRepository(db)

	// вспомогательные компоненты
	tokenProvider := services.NewTokenProvider(accessJwtMaker, refreshTokensRepo)
	emailProvider := services.NewEmailProvider(emailAuth)

	// сервисы
	authService := services.NewStudentsService(studentsRepo, uow)
	registerService := services.NewRegisterService(userRepo, studentsRepo, tokenProvider, uow)
	loginService := services.NewLoginService(userRepo, tokenProvider, uow)
	refreshService := services.NewRefreshTokensService(refreshTokensRepo, tokenProvider, uow)
	resetPasswordService := services.NewResetPasswordService(resetPasswordRepo, userRepo, emailProvider, uow)

	authGroup := r.Group("/auth")

	authGroup.POST("/check", auth.CheckStudent(authService))
	authGroup.POST("/register", auth.RegisterByToken(registerService))
	authGroup.POST("/login", auth.Login(loginService))
	authGroup.POST("/reset_password", auth.ResetPassword(resetPasswordService))
	authGroup.POST("/set_new_password", auth.SetNewPassword(resetPasswordService))
	authGroup.POST("/refresh", auth.RefreshToken(refreshService))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

package routes

import (
	"bobri/internal/api/controllers/users"
	"bobri/internal/api/repositories"
	"bobri/internal/api/services"
	"bobri/internal/middleware"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UserRoutes(r *gin.Engine, db *pgxpool.Pool, accessJWTMaker *helpers.JWTMaker) {
	uow := repositories.NewUoW(db)

	userHandlerGroup := r.Group("/me")
	userHandlerGroup.Use(middleware.AuthenticationMiddleware(accessJWTMaker, 10))

	// репозитории
	userRepo := repositories.NewUserRepository(db)
	completedEventRepo := repositories.NewCompletedEventsRepository(db)

	// сервисы
	userService := services.NewUserService(userRepo)
	completedEventService := services.NewCompletedEventsService(completedEventRepo, uow)

	// маршруты /me
	userHandlerGroup.GET("/profile", users.GetProfile(userService))
	userHandlerGroup.GET("/completed_events", users.GetCompletedEvents(completedEventService))

	// паблик маршрут
	r.GET("/leaderboard", users.GetLeaderboard(userService))
}

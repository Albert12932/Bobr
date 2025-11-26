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
	userHandlerGroup := r.Group("/me")

	completedEventRepo := repositories.NewCompletedEventsRepository(db)
	completedEventService := services.NewCompletedEventsService(completedEventRepo, db)
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	userHandlerGroup.Use(middleware.AuthenticationMiddleware(accessJWTMaker, 10))
	userHandlerGroup.GET("/profile", users.GetProfile(userService))
	userHandlerGroup.GET("/completed_events", users.GetCompletedEvents(completedEventService))

	r.GET("/leaderboard", users.GetLeaderboard(userService))
}

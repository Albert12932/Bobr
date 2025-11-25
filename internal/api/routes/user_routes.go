package routes

import (
	"bobri/internal/api/controllers/users"
	"bobri/internal/middleware"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UserRoutes(r *gin.Engine, db *pgxpool.Pool, accessJWTMaker *helpers.JWTMaker) {
	userHandlerGroup := r.Group("/me")
	userHandlerGroup.Use(middleware.AuthenticationMiddleware(accessJWTMaker, 10))
	userHandlerGroup.GET("/profile", users.GetProfile(db))
	userHandlerGroup.GET("/completed_events", users.GetCompletedEvents(db))
}

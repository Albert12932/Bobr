package routes

import (
	"bobri/internal/api/controllers/events"
	"bobri/internal/api/controllers/users"
	"bobri/internal/middleware"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AdminRoutes(r *gin.Engine, db *pgxpool.Pool, accessJWTMaker *helpers.JWTMaker) {
	// Создаем группу защищенных хэндлеров
	adminHandlersGroup := r.Group("/admin")
	adminHandlersGroup.Use(middleware.AuthenticationMiddleware(accessJWTMaker, 30))

	// users
	adminHandlersGroup.DELETE("/delete_user/:email", users.DeleteUser(db))
	adminHandlersGroup.GET("/students", users.GetStudents(db))
	adminHandlersGroup.GET("/users", users.GetUsers(db))
	adminHandlersGroup.PATCH("/update_user", users.UpdateUser(db))

	// events
	adminHandlersGroup.GET("/events", events.GetEvents(db))
	adminHandlersGroup.POST("/create_event", events.CreateEvent(db))
	adminHandlersGroup.PATCH("/update_event", events.UpdateEvent(db))
	adminHandlersGroup.DELETE("/delete_event/:id", events.DeleteEvent(db))
	adminHandlersGroup.POST("/add_completed_event", events.AddCompletedEvent(db))
	adminHandlersGroup.DELETE("/delete_completed_event/:user_id/:event_id", events.DeleteCompletedEvent(db))
	adminHandlersGroup.GET("/completed_events", events.GetAllCompletedEvents(db))
}

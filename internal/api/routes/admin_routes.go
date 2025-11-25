package routes

import (
	"bobri/internal/api/controllers/events"
	"bobri/internal/api/controllers/users"
	"bobri/internal/api/repositories"
	"bobri/internal/api/services"
	"bobri/internal/middleware"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AdminRoutes(r *gin.Engine, db *pgxpool.Pool, accessJWTMaker *helpers.JWTMaker) {
	// Создаем группу защищенных хэндлеров
	adminHandlersGroup := r.Group("/admin")
	adminHandlersGroup.Use(middleware.AuthenticationMiddleware(accessJWTMaker, 30))
	eventRepo := repositories.NewEventRepository(db)
	completedEventRepo := repositories.NewCompletedEventsRepository(db)
	completedEventService := services.NewCompletedEventsService(completedEventRepo, db)
	eventService := services.NewEventService(eventRepo)
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	studentRepo := repositories.NewStudentsRepository(db)
	studentService := services.NewStudentsService(studentRepo, db)

	// users
	adminHandlersGroup.DELETE("/delete_user/:email", users.DeleteUser(userService))
	adminHandlersGroup.GET("/students", users.GetStudents(studentService))
	adminHandlersGroup.GET("/users", users.GetUsers(userService))
	adminHandlersGroup.PATCH("/update_user", users.UpdateUser(userService))

	// events
	adminHandlersGroup.GET("/events", events.GetEvents(eventService))
	adminHandlersGroup.POST("/create_event", events.CreateEvent(eventService))
	adminHandlersGroup.PATCH("/update_event", events.UpdateEvent(eventService))
	adminHandlersGroup.DELETE("/delete_event/:id", events.DeleteEvent(eventService))
	adminHandlersGroup.POST("/add_completed_event", events.AddCompletedEvent(completedEventService))
	adminHandlersGroup.DELETE("/delete_completed_event/:user_id/:event_id", events.DeleteCompletedEvent(completedEventService))
	adminHandlersGroup.GET("/completed_events", events.GetAllCompletedEvents(completedEventService))
}

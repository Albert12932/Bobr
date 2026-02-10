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
	adminHandlersGroup := r.Group("/admin")
	adminHandlersGroup.Use(middleware.AuthenticationMiddleware(accessJWTMaker, 30))

	// создаем UoW
	uow := repositories.NewUoW(db)

	// репозитории
	eventRepo := repositories.NewEventRepository(db)
	completedEventRepo := repositories.NewCompletedEventsRepository(db)
	userRepo := repositories.NewUserRepository(db)
	studentRepo := repositories.NewStudentsRepository(db)

	// сервисы
	eventService := services.NewEventService(eventRepo, uow)
	completedEventService := services.NewCompletedEventsService(completedEventRepo, uow)
	userService := services.NewUserService(userRepo)
	studentService := services.NewStudentsService(studentRepo, uow)

	// users
	adminHandlersGroup.DELETE("/delete_user/:user_id", users.DeleteUser(userService))

	adminHandlersGroup.GET("/students", users.GetStudents(studentService))
	adminHandlersGroup.GET("/users", users.GetUsers(userService))
	adminHandlersGroup.PATCH("/update_user", users.UpdateUser(userService))

	// events
	adminHandlersGroup.GET("/events", events.GetEvents(eventService))
	adminHandlersGroup.POST("/create_event", events.CreateEvent(eventService))
	adminHandlersGroup.PATCH("/update_event", events.UpdateEvent(eventService))
	adminHandlersGroup.DELETE("/delete_event/:id", events.DeleteEvent(eventService))
	adminHandlersGroup.POST("/create_suggest", events.CreateSuggest(eventService))
	adminHandlersGroup.DELETE("/delete_suggestion/:id", events.DeleteSuggestion(eventService))

	// completed events
	adminHandlersGroup.POST("/add_completed_event", events.AddCompletedEvent(completedEventService))
	adminHandlersGroup.DELETE("/delete_completed_event/:user_id/:event_id", events.DeleteCompletedEvent(completedEventService))
	adminHandlersGroup.GET("/completed_events", events.GetAllCompletedEvents(completedEventService))
}

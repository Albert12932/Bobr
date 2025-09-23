package routes

import (
	"bobri/controllers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoutes(r *gin.Engine, db *pgxpool.Pool) {
	r.POST("/auth/check", controllers.AuthCheck(db))
}

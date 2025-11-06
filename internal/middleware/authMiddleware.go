package middleware

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthenticationMiddleware(accessJwtMaker *helpers.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Берем токен авторизации из header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				models.ErrorResponse{
					Error:   "Couldn't find Authorization header",
					Message: "Не удалось найти bearer Токен в header",
				})
			return
		}

		// Получаем часть с bearer токеном
		tokenString := strings.Split(authHeader, "Bearer ")[1]

		// Валидируем JWT токен
		_, err := accessJwtMaker.Verify(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось валидировать access токен",
			})
			return
		}

		// Продолжаем, если токен валиден
		c.Next()
	}
}

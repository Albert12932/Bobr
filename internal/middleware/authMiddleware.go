package middleware

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"fmt"
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
		claims, err := accessJwtMaker.Verify(tokenString)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось валидировать access токен. Возможно он истек",
			})
			return
		}

		// Просто создаем вашу модель прямо здесь
		payload := &models.Payload{}

		// Заполняем поля если они есть в claims
		if sub, ok := claims["sub"].(float64); ok {
			payload.Sub = int64(sub)
		}

		if roleLevel, ok := claims["roleLevel"].(float64); ok {
			payload.RoleLevel = int64(roleLevel)
		}

		if exp, ok := claims["exp"].(float64); ok {
			payload.Exp = int64(exp)
		}

		if iat, ok := claims["iat"].(float64); ok {
			payload.Iat = int64(iat)
		}

		fmt.Println(payload.RoleLevel)
		
		if payload.RoleLevel <= 50 {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Не достаточно прав",
			})
			return
		}

		// Сохраняем payload в модель и продолжаем, если токен валиден
		c.Set("userPayload", payload)
		c.Next()
	}
}

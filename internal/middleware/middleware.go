package middleware

import (
	"net/http"
	"strings"

	"go-simple-warehouse/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
)

func JWT(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenString := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}

func Error() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // run handlers

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			var details interface{} = err.Error()
			if ve, ok := err.(validator.ValidationErrors); ok {
				m := map[string]string{}
				for _, fe := range ve {
					field := strings.ToLower(fe.Field())
					if fe.Tag() == "oneof" {
						m[field] = "must be one of: " + fe.Param()
					} else {
						m[field] = fe.Tag()
					}
				}
				details = m
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request payload",
				"details": details,
			})
			c.Abort()
		}
	}
}

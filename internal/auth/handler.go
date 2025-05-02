package auth

import (
	"go-simple-warehouse/internal/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		var user User
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			c.Error(err)
			return
		}
		if err := user.CheckPassword(req.Password); err != nil {
			c.Error(err)
			return
		}
		// create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.ID,
			"exp": time.Now().Add(24 * time.Hour).Unix(), // 24 hours expiration
		})
		signed, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": signed})
	}
}

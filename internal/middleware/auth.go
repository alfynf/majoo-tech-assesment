package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/alfynf/simple-blog-api/internal/config"
	"github.com/alfynf/simple-blog-api/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format, expected 'Bearer <token>'")
)

// AuthMiddleware creates a gin.HandlerFunc for JWT authentication.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorResponse{Error: ErrMissingAuthHeader.Error()})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorResponse{Error: ErrInvalidAuthHeader.Error()})
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorResponse{Error: "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, err := uuid.Parse(claims["sub"].(string))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorResponse{Error: "Invalid user ID in token"})
				return
			}
			c.Set("userID", userID)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, handler.ErrorResponse{Error: "Invalid token"})
		}
	}
}

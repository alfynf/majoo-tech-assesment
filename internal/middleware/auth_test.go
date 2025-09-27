package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alfynf/simple-blog-api/internal/config"
	"github.com/alfynf/simple-blog-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createTestToken(userID uuid.UUID, secret string, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWTSecret: "test-secret"}
	userID := uuid.New()

	// Setup a test router with the middleware
	router := gin.New()
	router.Use(middleware.AuthMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		// This handler will only be reached if middleware passes
		contextUserID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, userID, contextUserID)
		c.Status(http.StatusOK)
	})

	t.Run("Success - Valid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		validToken, _ := createTestToken(userID, cfg.JWTSecret, time.Now().Add(time.Hour))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Failure - No Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "missing authorization header")
	})

	t.Run("Failure - Malformed Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "invalid-token-format")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid authorization header format")
	})

	t.Run("Failure - Expired Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		expiredToken, _ := createTestToken(userID, cfg.JWTSecret, time.Now().Add(-time.Hour))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", expiredToken))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("Failure - Invalid Signature", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		invalidToken, _ := createTestToken(userID, "wrong-secret", time.Now().Add(time.Hour))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", invalidToken))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("Failure - Invalid User ID in Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		// Create a token with a non-uuid subject
		claims := jwt.MapClaims{
			"sub": "not-a-uuid",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user ID in token")
	})
}

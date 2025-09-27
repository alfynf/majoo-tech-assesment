package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger returns a gin.HandlerFunc (middleware) that logs requests using zerolog.
// It logs the status code, method, path, latency, client IP, and user-agent.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		// Process the request
		c.Next()

		// After the request is handled, log the details
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Create a sub-logger with contextual fields
		event := log.With().
			Int("status_code", statusCode).
			Str("method", method).
			Str("path", path).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Logger()

		if errorMessage != "" {
			event.Error().Msg(errorMessage)
		} else {
			event.Info().Msg("request handled")
		}
	}
}

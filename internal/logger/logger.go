package logger

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"bierliste_backend/env"

	"github.com/gin-gonic/gin"
)

// Setup initialises the global slog logger based on environment config.
// Text format is used in development; JSON format everywhere else.
// Log level is read from LOG_LEVEL (default: info).
func Setup() {
	opts := &slog.HandlerOptions{Level: parseLevel(env.LogLevel.GetValue())}

	var handler slog.Handler
	if env.IsDevelopment() {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// RequestLogger returns a Gin middleware that logs each HTTP request via slog.
// Requests resulting in 5xx are logged at Error level, 4xx at Warn, others at Info.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if q := c.Request.URL.RawQuery; q != "" {
			path += "?" + q
		}

		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", time.Since(start),
			"ip", c.ClientIP(),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case status >= http.StatusInternalServerError:
			slog.Error("request", attrs...)
		case status >= http.StatusBadRequest:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}

// Recovery returns a Gin middleware that recovers from panics and logs them
// at Error level via slog before returning a 500 response.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		slog.Error("panic recovered",
			"error", err,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

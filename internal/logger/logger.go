package logger

import (
	"context"
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
//
// Error-level entries are always written to a JSON file in addition to stdout
// so that 500s and panics are persisted across process restarts. The file path
// is read from ERROR_LOG_FILE (default: errors.log).
func Setup() {
	opts := &slog.HandlerOptions{Level: parseLevel(env.LogLevel.GetValue())}

	var primary slog.Handler
	if env.IsDevelopment() {
		primary = slog.NewTextHandler(os.Stdout, opts)
	} else {
		primary = slog.NewJSONHandler(os.Stdout, opts)
	}

	errorLogPath := env.ErrorLogFile.GetValue()
	if errorLogPath == "" {
		errorLogPath = "errors.log"
	}

	errorFile, err := os.OpenFile(errorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Non-fatal: log to stdout only and continue.
		slog.New(primary).Warn("could not open error log file — errors will only be written to stdout",
			"path", errorLogPath,
			"error", err,
		)
		slog.SetDefault(slog.New(primary))
		return
	}

	// Error log always uses JSON regardless of environment so it is
	// machine-readable and easy to grep or ship to a log aggregator.
	errorFileHandler := slog.NewJSONHandler(errorFile, &slog.HandlerOptions{Level: slog.LevelError})

	slog.SetDefault(slog.New(multiHandler{primary, errorFileHandler}))
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

// multiHandler fans slog records out to multiple handlers.
// Each handler only receives records it is enabled for (e.g. the error file
// handler silently ignores Info/Warn records).
type multiHandler []slog.Handler

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make(multiHandler, len(m))
	for i, h := range m {
		handlers[i] = h.WithAttrs(attrs)
	}
	return handlers
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	handlers := make(multiHandler, len(m))
	for i, h := range m {
		handlers[i] = h.WithGroup(name)
	}
	return handlers
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

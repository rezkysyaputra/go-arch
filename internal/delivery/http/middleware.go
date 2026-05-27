package http

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		ctx.Set(requestIDKey, requestID)
		ctx.Writer.Header().Set(requestIDHeader, requestID)
		ctx.Next()
	}
}

func ZerologMiddleware(logger zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		status := ctx.Writer.Status()
		event := logger.Info()
		if len(ctx.Errors) > 0 {
			event = logger.Error().Str("gin_errors", ctx.Errors.String())
		} else if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}

		event.
			Str("request_id", requestIDFromContext(ctx)).
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Int("status", status).
			Dur("latency", time.Since(start)).
			Str("client_ip", ctx.ClientIP()).
			Str("user_agent", ctx.Request.UserAgent()).
			Msg("http request completed")
	}
}

func requestIDFromContext(ctx *gin.Context) string {
	requestID, ok := ctx.Get(requestIDKey)
	if !ok {
		return ""
	}

	value, ok := requestID.(string)
	if !ok {
		return ""
	}

	return value
}

func newRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}

	return hex.EncodeToString(bytes)
}

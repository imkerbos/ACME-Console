package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/ACME-Console/internal/logger"
	"go.uber.org/zap"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// SkipPaths is a list of paths to skip logging
var SkipPaths = map[string]bool{
	"/health": true,
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip logging for certain paths
		if SkipPaths[path] {
			c.Next()
			return
		}

		start := time.Now()
		query := c.Request.URL.RawQuery
		requestID := GetRequestID(c)

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = rw

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
		}

		// Add request body for non-GET requests (limit size)
		if c.Request.Method != "GET" && len(requestBody) > 0 && len(requestBody) < 4096 {
			fields = append(fields, zap.String("request_body", string(requestBody)))
		}

		// Add response body for errors (limit size)
		if status >= 400 && rw.body.Len() < 4096 {
			fields = append(fields, zap.String("response_body", rw.body.String()))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("Server error", fields...)
		case status >= 400:
			logger.Warn("Client error", fields...)
		default:
			logger.Info("Request", fields...)
		}
	}
}

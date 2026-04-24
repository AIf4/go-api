package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// continúa la cadena — handler se ejecuta aquí
		c.Next()

		// después del handler ya tienes el status y los errores
		latency := time.Since(start)
		status := c.Writer.Status()
		ginErrors := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// campos estructurados — fáciles de consultar en cualquier agregador
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", ip),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		}

		if ginErrors != "" {
			fields = append(fields, zap.String("errors", ginErrors))
		}

		// nivel del log según el status code
		switch {
		case status >= 500:
			log.Error("server error", fields...)
		case status >= 400:
			log.Warn("client error", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}

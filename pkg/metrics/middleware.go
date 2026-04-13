package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		HTTPRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(),strconv.Itoa(c.Writer.Status())).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(),strconv.Itoa(c.Writer.Status())).Observe(duration)
	}
}
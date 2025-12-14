package middleware

import (
	"fmt"
	"strings"
	"time"

	"my-project/logs"

	"github.com/gin-gonic/gin"
	//"my-project/logs/metrics" // Uncomment when metrics package is ready
)

func SetAPITimer() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Format Endpoint String (method.path_with_underscores)
		path := c.Request.URL.Path
		sanitizedPath := strings.ReplaceAll(path, "/", "_")
		// Remove leading underscore if present for cleaner metrics
		sanitizedPath = strings.TrimPrefix(sanitizedPath, "_")
		endpoint := fmt.Sprintf("%s.%s", c.Request.Method, sanitizedPath)

		// 2. Metrics: Increment Call (Commented until metrics package exists)
		logs.Client.Increment("api.calls." + endpoint)

		// 3. Log Incoming Request
		// We format the metadata into a string since our basic logger might not support objects yet
		logs.Info(fmt.Sprintf("Incoming request | Method: %s | Path: %s | IP: %s | UA: %s",
			c.Request.Method,
			path,
			c.ClientIP(),
			c.Request.UserAgent(),
		))

		// 4. Process Request (Pass to next middleware/controller)
		c.Next()

		// 5. Post-Processing (After response is sent)
		duration := time.Since(start)
		durationMs := float64(duration.Milliseconds())
		statusCode := c.Writer.Status()

		// 6. Metrics: Timing (Commented until metrics package exists)
		logs.Client.Timing("api.latency."+endpoint, durationMs)

		// 7. Determine Log Level based on Status Code
		logMessage := fmt.Sprintf("Request finished in %vms | Status: %d | Method: %s | Path: %s",
			durationMs, statusCode, c.Request.Method, path)

		if statusCode >= 500 {
			logs.Error(logMessage)
		} else if statusCode >= 400 {
			logs.Warn(logMessage)
		} else {
			logs.Info(logMessage)
		}
	}
}

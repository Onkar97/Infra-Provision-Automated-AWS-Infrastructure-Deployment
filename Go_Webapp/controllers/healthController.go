package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"my-project/db"
	"my-project/logs"
	"my-project/models"
)

// GetHealth handles the health check endpoint
func GetHealth(c *gin.Context) {
	// 1. Method Check (Gin usually handles this via routing, but we can double check)
	if c.Request.Method != "GET" {
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	// 2. Strict Validation
	// Check for body (Content-Length or Transfer-Encoding), Query Params, or Auth Header
	hasBody := c.Request.ContentLength > 0 || c.Request.Header.Get("Transfer-Encoding") != ""
	hasQueryString := len(c.Request.URL.Query()) > 0
	hasAuthHeader := c.Request.Header.Get("Authorization") != ""

	if hasBody || hasQueryString || hasAuthHeader {
		c.Status(http.StatusBadRequest)
		return
	}

	// 3. Database Operation (Insert Empty Record)
	// --- Start Timer Block ---
	startInsert := time.Now()

	healthCheck := models.HealthCheck{
		// CheckID is auto-increment, CheckDatetime defaults to NOW()
	}

	// Equivalent to: .insert().into(Health_Checks).values({})
	if err := db.DB.Create(&healthCheck).Error; err != nil {
		logs.Error("Health insert failed: " + err.Error())
		
		// 503 Service Unavailable for DB errors
		c.Status(http.StatusServiceUnavailable)
		return
	}

	insertDurationMs := float64(time.Since(startInsert).Milliseconds())
	logs.Info("Query executed in " + strconv.FormatFloat(insertDurationMs, 'f', 2, 64) + "ms")
	
	// metricsClient.Timing("db.query.latency.insertHealthCheck", insertDurationMs) // Uncomment when metrics are ready

	logs.Info("Assignment 9 Health Check Successful")

	// 4. Success Response
	c.Status(http.StatusOK)
}
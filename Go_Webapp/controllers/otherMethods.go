package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OtherMethods handles unsupported HTTP methods by returning 405
func OtherMethods(c *gin.Context) {
	// Equivalent to: return res.status(405).end();
	c.Status(http.StatusMethodNotAllowed)
}
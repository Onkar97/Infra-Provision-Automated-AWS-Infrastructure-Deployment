package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"my-project/db"
	"my-project/logs"
	"my-project/models"
)

// AuthenticateUser middleware handles Basic Authentication
func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get Basic Auth credentials from the standard library helper
		username, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth {
			c.Header("WWW-Authenticate", "Basic")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 2. Find User in DB
		var user models.User
		// Equivalent to: .where("user.username = :username", { username }).getOne()
		if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
			logs.Info("Cannot find User: " + username)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 3. Compare Password
		// bcrypt.CompareHashAndPassword returns nil on success
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			logs.Info("Password does not match for user: " + username)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 4. Attach User to Context
		// This is critical: It allows c.Get("user") to work in your controllers
		c.Set("user", &user)

		// 5. Continue to the next handler
		c.Next()
	}
}
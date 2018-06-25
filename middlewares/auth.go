package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Auth middleware set current user by given cookie token.
func Auth() func(c *gin.Context) {
	return func(c *gin.Context) {
		if os.Getenv("ENV") == "development" {
			c.Next()
			return
		}

		token, err := c.Cookie("token")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "can not auth, please enable cookie.",
			})
			return
		}

		if token == os.Getenv("KALAN_TOKEN") {
			c.Set("current_user", "kalan")
			c.Next()
			return
		}

		if token == os.Getenv("YVETTE_TOKEN") {
			c.Set("current_user", "yvette")
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "you have no permission to do this action.",
		})
	}
}

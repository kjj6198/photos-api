package apis

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

type authInput struct {
	Token string `json:"token"`
}

func auth(c *gin.Context) {
	input := &authInput{}
	c.ShouldBindJSON(input)

	fmt.Println(c.Get("current_user"))
	if input.Token != "" {
		switch input.Token {
		case os.Getenv("KALAN_TOKEN"):
			c.SetCookie("token", os.Getenv("KALAN_TOKEN"), 60*60*24*30, "", "", false, true)
		case os.Getenv("YVETTE_TOKEN"):
			c.SetCookie("token", os.Getenv("YVETTE_TOKEN"), 60*60*24*30, "", "", false, true)
		}
		c.JSON(200, gin.H{"message": "ok"})
		return
	}

	c.JSON(400, gin.H{
		"message": "can not auth.",
	})
}

func RegisterAuthHandler(router *gin.RouterGroup) {
	router.POST("/", auth)
}

package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kjj6198/photos-api/apis"
	"github.com/kjj6198/photos-api/config"
	"github.com/kjj6198/photos-api/db"
	"github.com/kjj6198/photos-api/middlewares"
	"github.com/kjj6198/photos-api/services"
)

func main() {
	config.Load()

	router := gin.Default()
	db := db.NewDB()

	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("uploader", services.NewUploader(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
		))
	})

	router.Use(middlewares.Auth())

	if os.Getenv("env") == "development" {
		router.StaticFile("/", "./index.html")
	}

	v1 := router.Group("api/v1")
	router.NoRoute(func(c *gin.Context) {
		if c.GetHeader("Content-Type") == "application/json" {
			c.AbortWithStatusJSON(404, nil)
			return
		}
		c.Next()
	})

	apis.RegisterWorkHandler(v1.Group("/works"))
	apis.RegisterAuthHandler(v1.Group("/auth"))

	router.Run()
}

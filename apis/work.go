package apis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"github.com/kjj6198/photos-api/models"
	"github.com/kjj6198/photos-api/services"
	"github.com/kjj6198/photos-api/utils"
)

func getWorks(c *gin.Context) {
	ctx := context.Background()
	valueCtx := context.WithValue(ctx, "db", c.MustGet("db"))

	work := &models.Work{}

	cursor := aws.String("")
	works, nextCursor, err := work.FindMany(valueCtx, 1, *cursor)
	if err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"message": err,
		})
	}

	if nextCursor != "" {
		c.JSON(200, gin.H{
			"works":  works,
			"cursor": nextCursor,
		})
		return
	}

	c.JSON(200, gin.H{
		"works": works,
	})

}

func getWork(c *gin.Context) {
	id := c.Param("id")
	if id != "" {
		work := &models.Work{
			ID: id,
		}

		ctx := context.Background()
		valueCtx := context.WithValue(ctx, "db", c.MustGet("db"))
		work, err := work.FindOne(valueCtx)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

			return
		}

		c.JSON(200, work)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"message": "params `id` must exist.",
	})

	return
}

type workInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func createWork(c *gin.Context) {
	input := &workInput{}
	err := c.ShouldBindJSON(input)
	if err != nil {
		c.JSON(400, gin.H{
			"message": fmt.Sprintln("can not create work.", err),
		})
		return
	}
	work := &models.Work{
		ID:          utils.GenerateUUID(),
		Name:        input.Name,
		Description: input.Description,
		Author:      "kalan",
		CreatedAt:   time.Now(),
	}

	ctx := context.Background()
	valueCtx := context.WithValue(ctx, "db", c.MustGet("db"))
	err = work.Create(valueCtx)
	if err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"message": fmt.Sprint(err),
		})
		return
	}

	c.JSON(200, work)
}

type uploadInfo struct {
	WorkID   string
	Filename string
	Type     string
	Data     []byte
}

func upload(uploader *services.Uploader, info *uploadInfo, c chan models.Image) {

	fileInfo, err := uploader.Upload(
		info.WorkID,
		info.Filename,
		"image/jpeg",
		info.Data,
	)
	fmt.Println("uploading file to s3....")
	if err != nil {
		fmt.Println(err)
	}

	imageInfo := &models.ImageInfo{}
	imageInfo, err = imageInfo.GetImageInfo(info.Data)

	if err != nil {
		fmt.Println(err)
		return
	}

	c <- models.Image{
		ImageURL: &models.ImageURL{
			Original: fileInfo.URL,
		},
		ImageInfo: imageInfo,
	}
}

func createWorkImages(c *gin.Context) {

	multipart, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{
			"message": "can not read multipart data.",
		})
		return
	}

	files := multipart.File["files"]

	images := []models.Image{}

	for _, file := range files {
		fmt.Println(file)
		fmt.Println("hello world")
	}

	fmt.Println("lock end")
	c.JSON(200, images)
}

func getWorkImages(c *gin.Context) {

}

func RegisterWorkHandler(router *gin.RouterGroup) {
	router.GET("/:id", getWork)
	router.POST("/:id/images", createWorkImages)
	router.GET("/", getWorks)
	router.POST("/", createWork)
}

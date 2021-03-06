package apis

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/kjj6198/photos-api/models"
	"github.com/kjj6198/photos-api/services"
	"github.com/kjj6198/photos-api/utils"
)

type getWorksInput struct {
	Limit  int64  `form:"limit"`
	Cursor string `form:"cursor"`
}

func (input *getWorksInput) GetLimit() int64 {
	if input.Limit == 0 {
		return 100
	}

	return input.Limit
}

func getWorks(c *gin.Context) {
	ctx := context.Background()
	valueCtx := context.WithValue(ctx, "db", c.MustGet("db"))
	input := &getWorksInput{}
	c.ShouldBindQuery(input)
	fmt.Println(input)

	work := &models.Work{}

	works, nextCursor, err := work.FindMany(valueCtx, input.GetLimit(), input.Cursor)
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

	c.JSON(200, works)
}

func getWork(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*dynamodb.DynamoDB)
	if id != "" {
		work := &models.Work{
			ID: id,
		}

		ctx := context.Background()
		valueCtx := context.WithValue(ctx, "db", db)
		work, err := work.FindOne(valueCtx)

		img := &models.Image{
			WorkID: id,
		}

		images, _, _ := img.GetImages(db, "")
		fmt.Println(images)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "can not find work.",
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
	Wg       *sync.WaitGroup
}

func upload(uploader *services.Uploader, info *uploadInfo, c chan models.Image) {
	defer info.Wg.Done()
	fileInfo, err := uploader.Upload(
		info.WorkID,
		info.Filename,
		"image/jpeg",
		info.Data,
	)

	if err != nil {
		fmt.Println(err)
	}

	imageInfo := &models.ImageInfo{}
	imageInfo, err = imageInfo.GetImageInfo(info.Data)

	if err != nil {
		fmt.Println(err)
	}

	c <- models.Image{
		WorkID: info.WorkID,
		ImageURL: &models.ImageURL{
			Original: fileInfo.URL,
		},
		ImageInfo: imageInfo,
	}
}

func createWorkImages(c *gin.Context) {
	db := c.MustGet("db").(*dynamodb.DynamoDB)
	workID := c.Param("id")
	multipart, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{
			"message": "can not read multipart data.",
		})
		return
	}

	files := multipart.File["files"]

	images := []models.Image{}
	receiver := make(chan models.Image, len(files))
	wg := &sync.WaitGroup{}

	for _, file := range files {
		wg.Add(1)
		f, _ := file.Open()
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("error during uploading file, skip.")
		}

		go upload(c.MustGet("uploader").(*services.Uploader), &uploadInfo{
			WorkID:   workID,
			Filename: file.Filename,
			Type:     "image/jpeg",
			Data:     data,
			Wg:       wg,
		}, receiver)
	}

	wg.Wait()
	close(receiver)
	for val := range receiver {
		images = append(images, val)
	}

	ctx := context.Background()
	valueCtx := context.WithValue(ctx, "db", db)

	for _, img := range images {
		output, _ := img.CreateImage(valueCtx)
		fmt.Println(output)
	}
	c.JSON(200, images)
}

func getWorkImages(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*dynamodb.DynamoDB)
	img := &models.Image{
		WorkID: id,
	}

	images, _, err := img.GetImages(db, "")
	if err != nil {
		fmt.Println(err)
	}

	c.JSON(200, images)
}

func RegisterWorkHandler(router *gin.RouterGroup) {
	router.GET("/:id", getWork)
	router.GET("/:id/images", getWorkImages)
	router.POST("/:id/images", createWorkImages)
	router.GET("/", getWorks)

	router.POST("/", createWork)
}

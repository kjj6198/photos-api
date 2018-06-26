package models

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/satori/go.uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/rwcarlsen/goexif/exif"
)

type ImageURL struct {
	Original string `json:"original,omitempty"`
	Normal   string `json:"normal"`
}

type ImageInfo struct {
	Make     string `json:"make"`      // 相機製造商
	Model    string `json:"model"`     // 機型
	Exposure string `json:"exposure"`  // 曝光時間
	Aperture string `json:"aperture"`  // 光圈值
	FocalLen string `json:"focal_len"` // 焦距
	Author   string `json:"author"`
}

type Image struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	ImageURL  *ImageURL  `json:"image_url"`
	ImageInfo *ImageInfo `json:"image_info"`
	WorkID    string     `json:"-"`
}

func (img *Image) GetImagesCount(db *dynamodb.DynamoDB) (int64, error) {
	filter := expression.Name("work_id").Equal(expression.Value(img.WorkID))
	exp, _ := expression.NewBuilder().WithFilter(filter).Build()
	input := &dynamodb.ScanInput{
		TableName:                 aws.String("images"),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		FilterExpression:          exp.Filter(),
	}

	output, err := db.Scan(input)
	if err != nil {
		return 0, err
	}

	return *output.ScannedCount, nil
}

func (img *Image) GetImages(db *dynamodb.DynamoDB, cursor string) (images []*Image, nextCursor string, err error) {
	filter := expression.Name("work_id").Equal(expression.Value(img.WorkID))
	exp, _ := expression.NewBuilder().WithFilter(filter).Build()
	var exclusiveKey map[string]*dynamodb.AttributeValue

	if cursor != "" {
		exclusiveKey = map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(cursor),
			},
		}
	}

	input := &dynamodb.ScanInput{
		TableName: aws.String("images"),
		Limit:     aws.Int64(100), // TODO: cursor base
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		FilterExpression:          exp.Filter(),
		ExclusiveStartKey:         exclusiveKey,
	}

	output, err := db.Scan(input)
	if err != nil {
		return nil, "", err
	}

	dynamodbattribute.UnmarshalListOfMaps(output.Items, &images)
	if output.LastEvaluatedKey != nil {
		return images, *output.LastEvaluatedKey["id"].S, nil
	}

	return images, "", nil
}

func (img *Image) FindOne(db *dynamodb.DynamoDB, imageID string) (result *Image, err error) {
	output, err := db.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(imageID),
			},
		},
	})

	if err != nil {
		return nil, err
	}
	result = &Image{}
	dynamodbattribute.UnmarshalMap(output.Item, result)

	return result, nil
}

func trim(str string) string {
	return strings.Replace(str, "\"", "", -1)
}

// CreateImage creates image by Image struct
func (img *Image) CreateImage(ctx context.Context) (output *dynamodb.PutItemOutput, err error) {
	db := ctx.Value("db").(*dynamodb.DynamoDB)

	id, _ := uuid.NewV4()
	img.ID = id.String()
	img.CreatedAt = time.Now()

	item, _ := dynamodbattribute.MarshalMap(img)

	output, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("images"),
		Item:      item,
	})

	if err != nil {
		return nil, err
	}

	return output, err
}

// GetImageInfo returns exif info by given data.
func (info *ImageInfo) GetImageInfo(data []byte) (result *ImageInfo, err error) {
	reader := bytes.NewReader(data)
	x, err := exif.Decode(reader)
	if err != nil {
		return nil, err
	}

	make, _ := x.Get(exif.Make)
	model, _ := x.Get(exif.Model)
	exposure, _ := x.Get(exif.ExposureTime)
	aperture, _ := x.Get(exif.ApertureValue)
	focal, _ := x.Get(exif.FocalLength)
	author, _ := x.Get(exif.Artist)

	return &ImageInfo{
		Make:     trim(make.String()),
		Model:    trim(model.String()),
		Exposure: trim(exposure.String()),
		Aperture: trim(aperture.String()),
		FocalLen: trim(focal.String()),
		Author:   trim(author.String()),
	}, nil
}

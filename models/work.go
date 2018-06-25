package models

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/photos/utils"
)

// Work represents for "作品集"
type Work struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Position    string    `json:"position"`
	Lat         float64   `json:"lat,omitempty"`
	Lng         float64   `json:"lng,omitempty"`
	Images      []*Image  `json:"images,omitempty"`
	Author      string    `json:"author"`
	Cover       *Image    `json:"cover,omitempty"`
}

// FindOne finds a work
func (work *Work) FindOne(ctx context.Context) (result *Work, err error) {
	db := ctx.Value("db").(*dynamodb.DynamoDB)
	res, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("works"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(work.ID),
			},
		},
	})

	if err != nil {
		return nil, err
	}
	result = &Work{}
	dynamodbattribute.UnmarshalMap(res.Item, result)
	image := &Image{
		WorkID: result.ID,
	}
	images, _, err := image.GetImages(db)
	if err != nil {
		return nil, err
	}
	result.Images = images

	return result, nil
}

// FindMany finds many works.
func (work *Work) FindMany(
	ctx context.Context,
	limit int64,
	cursor string,
) (result []*Work, nextCursor string, err error) {
	db := ctx.Value("db").(*dynamodb.DynamoDB)

	limit = int64(math.Max(1, math.Min(float64(limit), 100.0)))
	av, err := utils.Base64ToDynamoDBAttributeValue(cursor)

	if err != nil || len(av) == 0 {
		fmt.Println("can not convert cursor, skip it.")
		av = nil
	}

	input := &dynamodb.ScanInput{
		TableName:         aws.String("works"),
		Limit:             aws.Int64(limit),
		ExclusiveStartKey: av,
	}

	output, err := db.Scan(input)

	if err != nil {
		fmt.Println("can not find works.", err)
		return nil, "", err
	}

	dynamodbattribute.UnmarshalListOfMaps(output.Items, &result)
	nextCursor = utils.DynamoDBAttributeValueToBase64(output.LastEvaluatedKey)
	return result, nextCursor, nil
}

// Create creates a work in dynamoDB
func (work *Work) Create(ctx context.Context) error {
	db := ctx.Value("db").(*dynamodb.DynamoDB)
	av, err := dynamodbattribute.MarshalMap(work)

	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("works"),
		Item:      av,
	}

	_, err = db.PutItem(input)

	if err != nil {
		return err
	}

	return nil
}

// Update updates value
func (work *Work) Update(ctx context.Context) error {

	return fmt.Errorf("not implemented")
}

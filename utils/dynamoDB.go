package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Base64ToDynamoDBAttributeValue converts base64 to real string
func Base64ToDynamoDBAttributeValue(str string) (map[string]*dynamodb.AttributeValue, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(err)
	}

	data := make(map[string]interface{})

	json.Unmarshal(dataBytes, &data)
	return dynamodbattribute.MarshalMap(data)
}

// DynamoDBAttributeValueToBase64 converts attributes to base64 string
func DynamoDBAttributeValueToBase64(in map[string]*dynamodb.AttributeValue) string {
	if len(in) == 0 {
		return ""
	}

	data, _ := json.Marshal(in)

	return base64.StdEncoding.EncodeToString(data)
}

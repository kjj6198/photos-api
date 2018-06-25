package services

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Uploader struct {
	s3 *s3.S3
}

type fileInfo struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

func (uploader *Uploader) Upload(
	folder string,
	filename string,
	contentType string,
	body []byte,
) (info *fileInfo, err error) {
	bucket := os.Getenv("AWS_BUCKET")

	_, err = uploader.s3.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
		Key:         aws.String(fmt.Sprintf("%s/%s", folder, filename)),
		Body:        bytes.NewReader(body),
	})

	if err != nil {
		return nil, err
	}

	info = &fileInfo{}
	info.URL = fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("S3_URL"),
		bucket,
		fmt.Sprintf("%s/%s", folder, filename),
	)

	info.Filename = fmt.Sprintf("%s/%s", folder, filename)

	return info, nil
}

func NewUploader(awsAccessKeyID string, awsSecretAccessKey string) *Uploader {
	creds := credentials.NewStaticCredentials(
		awsAccessKeyID,
		awsSecretAccessKey,
		"",
	)

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-1"),
		Credentials: creds,
	})

	svc := s3.New(sess)

	return &Uploader{
		s3: svc,
	}
}

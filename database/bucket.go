package database

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type S3PutObject interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Configuration struct {
	Key      string
	Secret   string
	Token    string
	Endpoint string
	Region   string
}

// GetS3Client accepts an S3Configuration parameter and
// attempts to build a new S3 Client instance.
func GetS3Client(params *S3Configuration) (*s3.Client, error) {
	creds := credentials.NewStaticCredentialsProvider(
		params.Key,
		params.Secret,
		"",
	)

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           params.Endpoint,
			SigningRegion: params.Region,
		}, nil
	})

	conf, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(creds))

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(conf)
	return client, nil
}

// UploadFile accepts an S3 client instance, a bucket name, and
// a file buffer to send to the S3 instance.
//
// The file type is automatically determined by AWS SDK
//
// If successful, the filename (UUID) will be returned
func UploadFile(s3Client *s3.Client, bucket string, file []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	id := uuid.New()

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(id.String()),
		Body:        bytes.NewReader(file),
		ContentType: aws.String(http.DetectContentType(file)),
	}

	// TODO: Determine if we want to use the returned metadata for anything here
	_, err := s3Client.PutObject(ctx, input)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

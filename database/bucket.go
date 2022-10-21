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
// # The file type is automatically determined by AWS SDK
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

// SignUrl generates a new temporary pre-signed url that grants
// a user access to view content within the provided bucket
func SignUrl(s3Client *s3.Client, bucket string, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	presignClient := s3.NewPresignClient(s3Client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	}

	duration := func(po *s3.PresignOptions) {
		po.Expires = 30 * time.Second
	}

	presignResult, err := presignClient.PresignGetObject(ctx, input, duration)
	return presignResult.URL, err
}

// Exists accepts an S3 client instance, a bucket name, and
// a file name to check if it exists within the S3 instance.
//
// If successful, a bool of true will be returned
func Exists(s3Client *s3.Client, bucket string, filename string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	}

	_, err := s3Client.HeadObject(ctx, input)
	if err != nil {
		return false, err
	}

	return true, nil
}

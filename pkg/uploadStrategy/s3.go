package uploadStrategy

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appCfg "github.com/saleh-ghazimoradi/GopherMarket/config"
)

type S3Strategy struct {
	client         *s3.Client
	transferClient *transfermanager.Client
	bucket         string
	endpoint       string
	cfg            *appCfg.Config
}

func (s *S3Strategy) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer src.Close()

	result, err := s.transferClient.UploadObject(context.TODO(), &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   src,
	})
	if err != nil {
		return "", fmt.Errorf("uploading file: %w", err)
	}

	return *result.Key, nil
}

func (s *S3Strategy) DeleteFile(path string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.TrimPrefix(path, "/")),
	})
	return err
}

func NewS3Strategy(cfg *appCfg.Config) *S3Strategy {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWS.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKeyId,
			cfg.AWS.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		panic("failed to create AWS config " + err.Error())
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.AWS.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.AWS.S3Endpoint)
			o.UsePathStyle = true
		}
	})

	transferClient := transfermanager.New(client)

	return &S3Strategy{
		client:         client,
		transferClient: transferClient,
		bucket:         cfg.AWS.S3Bucket,
		endpoint:       cfg.AWS.S3Endpoint,
	}
}

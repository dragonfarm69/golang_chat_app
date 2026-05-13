package storage

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	s3_client           *s3.Client
	s3_presigned_client *s3.PresignClient
}

func NewCloudService() (*S3Service, error) {
	//load cloud storage
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("minioadmin", "miniocloud", "")),
	)

	if err != nil {
		log.Fatal("cannot load sdk config: ", err)
		return nil, err
	}

	s3_client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://localhost:9000")
		o.UsePathStyle = true
	})
	log.Println("Created s3 client")

	presignedClient := s3.NewPresignClient(s3_client)
	log.Println("minIO should work now")

	return &S3Service{
		s3_client:           s3_client,
		s3_presigned_client: presignedClient,
	}, nil
}

func (service *S3Service) GeneratePutPresignedURL(ctx context.Context, bucket string, key string, content_type string) (string, error) {
	req, err := service.s3_presigned_client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: &content_type,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(10 * time.Minute)
	})

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (service *S3Service) GenerateGetPresignedURL(ctx context.Context, bucket string, key string) (string, error) {
	req, err := service.s3_presigned_client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(10 * time.Minute)
	})

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (service *S3Service) GenerateDeletePresignURL(ctx context.Context, bucket string, key string) (string, error) {
	req, err := service.s3_presigned_client.PresignDeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(10 * time.Minute)
	})

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

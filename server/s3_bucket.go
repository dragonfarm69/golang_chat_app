package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (app *App) generatePutPresignedURL(ctx context.Context, bucket string, key string) (string, error) {
	req, err := app.s3_presigned_client.PresignPutObject(ctx, &s3.PutObjectInput{
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

func (app *App) generateGetPresignedURL(ctx context.Context, bucket string, key string) (string, error) {
	req, err := app.s3_presigned_client.PresignGetObject(ctx, &s3.GetObjectInput{
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

func (app *App) generateDeletePresignURL(ctx context.Context, bucket string, key string) (string, error) {
	req, err := app.s3_presigned_client.PresignDeleteObject(ctx, &s3.DeleteObjectInput{
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

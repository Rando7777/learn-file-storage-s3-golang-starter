package main

import (
	"time"
	"fmt"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error){
	presignClient := *s3.NewPresignClient(s3Client)
	
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	presignedReq, err := presignClient.PresignGetObject(
		context.TODO(),
		input,
		s3.WithPresignExpires(expireTime),
	)
	if err != nil {
		return "", fmt.Errorf("Error generating presigned url: %s", err)
	}
	return presignedReq.URL, nil
}

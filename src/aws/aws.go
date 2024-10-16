package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appConfig "kadane.xyz/go-backend/v2/src/config"
)

func NewAWSClient(appConfig *appConfig.Config) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-2"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			appConfig.AWSKey,
			appConfig.AWSSecret,
			"",
		)),
	)
	if err != nil {
		log.Printf("error loading AWS config: %v\n", err)
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	if client == nil {
		log.Printf("error creating AWS client: %v\n", err)
		return nil, err
	}

	return client, nil
}

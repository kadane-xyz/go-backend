package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	serverConfig "kadane.xyz/go-backend/v2/internal/config"
)

func NewAWSClient(cfg *serverConfig.Config) (*s3.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWS.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.Key,
			cfg.AWS.Secret,
			"",
		)),
	)
	if err != nil {
		log.Printf("error loading AWS config: %v\n", err)
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)
	if client == nil {
		log.Printf("error creating AWS client: %v\n", err)
		return nil, err
	}

	return client, nil
}

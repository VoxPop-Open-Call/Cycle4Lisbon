package aws

import (
	"context"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type Client struct {
	S3  *S3
	SES *SES
}

type Config struct {
	Region          string
	AccessKeyId     string
	SecretAccessKey string
	BucketName      string
	SESDomain       string
	SESFromName     string
	APIScheme       string
	APIEndpoint     string
}

func New(config Config) (*Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(config.Region),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				config.AccessKeyId,
				config.SecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	return &Client{
		newS3(cfg, config.BucketName),
		newSES(cfg, config.SESDomain, config.SESFromName, config.APIScheme, config.APIEndpoint),
	}, nil
}

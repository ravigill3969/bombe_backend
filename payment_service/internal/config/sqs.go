package config

import (
	"context"

	sqscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func InitSQS() *sqs.Client {
	cfg, err := sqscfg.LoadDefaultConfig(context.TODO(), sqscfg.WithRegion("us-east-1"))
	if err != nil {
		panic(err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	return sqsClient
}

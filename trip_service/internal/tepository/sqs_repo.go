package repository

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSQueue struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSQueue(client *sqs.Client, queueURL string) *SQSQueue {

	return &SQSQueue{
		client:   client,
		queueURL: queueURL,
	}
}

func (q *SQSQueue) ReceiveMessages(
	ctx context.Context,
) ([]types.Message, error) {

	result, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})

	if err != nil {
		return nil, err
	}

	return result.Messages, nil
}

func (q *SQSQueue) DeleteMessage(
	ctx context.Context,
	receiptHandle string,
) error {

	_, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})

	return err
}

func (q *SQSQueue) SendMessage(
	ctx context.Context,
	data any,

) error {

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:     aws.String(q.queueURL),
		MessageBody:  aws.String(string(body)),
		DelaySeconds: 0,
	})

	return err
}

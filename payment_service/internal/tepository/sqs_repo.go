package repository

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSQueue struct {
	client   *sqs.Client
	queueURL string
}

type SQSData struct {
	Status     int32  `json:"status"`
	Aud        string `json:"aud"`
	TempRideID string `json:"temp_ride_id"`
	PaymentID  string `json:"payment_id"`
	RiderID    string `json:"rider_id"`
	ServerType string `json:"server_type"`
	Message    string `json:"message"`

	Pickup      Location    `json:"pickup"`
	Dropoff     Location    `json:"dropoff"`
	Fare        Fare        `json:"fare"`
	RideDetails RideDetails `json:"ride_details"`
}

type Location struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Fare struct {
	AmountInCents int64  `json:"amount_in_cents"`
	Currency      string `json:"currency"`
}

type RideDetails struct {
	DistanceMeters  int64  `json:"distance_meters"`
	DurationSeconds int64  `json:"duration_seconds"`
	ServiceType     string `json:"service_"`
}

func NewSQSQueue(client *sqs.Client, queueURL string) *SQSQueue {

	return &SQSQueue{
		client:   client,
		queueURL: queueURL,
	}
}

func (q *SQSQueue) PublishRideRequest(
	ctx context.Context,
	data SQSData,
) error {

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.queueURL),
		MessageBody: aws.String(string(body)),
	})

	return err
}

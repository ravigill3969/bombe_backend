package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"payment_service/internal/model"
	repository "payment_service/internal/tepository"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/refund"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentService struct {
	paymentRepo repository.PaymentRepo
	sqsQueue    *repository.SQSQueue
}

func NewPaymentService(paymentRepo repository.PaymentRepo, sqsQueue *repository.SQSQueue) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		sqsQueue:    sqsQueue,
	}
}

func (p *PaymentService) CreateCheckoutSession(ctx context.Context, params model.CreateCheckoutParams, riderID string) (*model.CheckoutSessionResult, error) {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return nil, status.Error(codes.Internal, "STRIPE_SECRET_KEY is not set")
	}
	riderFrontendURL := os.Getenv("RIDER_FRONTEND_URL")
	if riderFrontendURL == "" {
		return nil, status.Error(codes.Internal, "RIDER_FRONTEND_URL is not set")
	}

	tempRideID := uuid.NewString()

	sc := stripe.NewClient(stripeKey)

	stripeParams := &stripe.CheckoutSessionCreateParams{
		SuccessURL: stripe.String(fmt.Sprintf("%s/rider/rider-active-trip", riderFrontendURL)),
		CancelURL:  stripe.String(fmt.Sprintf("%s/cancel", riderFrontendURL)),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
					Currency:   stripe.String(params.Fare.Currency),
					UnitAmount: stripe.Int64(params.Fare.AmountInCents),
					ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("%s Ride", params.ServiceType)),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"pickup_location":  params.Pickup.Address,
			"dropoff_location": params.Dropoff.Address,
			"pickup_lat":       fmt.Sprintf("%f", params.Pickup.Latitude),
			"pickup_lng":       fmt.Sprintf("%f", params.Pickup.Longitude),
			"dropoff_lat":      fmt.Sprintf("%f", params.Dropoff.Latitude),
			"dropoff_lng":      fmt.Sprintf("%f", params.Dropoff.Longitude),
			"amount":           fmt.Sprintf("%d", params.Fare.AmountInCents),
			"service_type":     params.ServiceType,
			"rider_id":         riderID,
			"temp_ride_id":     tempRideID,
			"distance_meters":  fmt.Sprintf("%d", params.Ride.DistanceMeters),
			"duration_seconds": fmt.Sprintf("%d", params.Ride.DurationSeconds),
		},

		PaymentIntentData: &stripe.CheckoutSessionCreatePaymentIntentDataParams{
			Metadata: map[string]string{
				"pickup_location":  params.Pickup.Address,
				"dropoff_location": params.Dropoff.Address,
				"pickup_lat":       fmt.Sprintf("%f", params.Pickup.Latitude),
				"pickup_lng":       fmt.Sprintf("%f", params.Pickup.Longitude),
				"dropoff_lat":      fmt.Sprintf("%f", params.Dropoff.Latitude),
				"dropoff_lng":      fmt.Sprintf("%f", params.Dropoff.Longitude),
				"amount":           fmt.Sprintf("%d", params.Fare.AmountInCents),
				"service_type":     params.ServiceType,
				"rider_id":         riderID,
				"temp_ride_id":     tempRideID,
				"distance_meters":  string(params.Ride.DistanceMeters),
				"duration_seconds": string(params.Ride.DurationSeconds),
			},
		},

		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}
	stripeParams.SetIdempotencyKey(fmt.Sprintf("checkout-%s-%d", riderID, time.Now().UnixNano()))

	session, err := sc.V1CheckoutSessions.Create(ctx, stripeParams)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create Stripe checkout session: %v", err))
	}

	_, err = p.paymentRepo.CreateStripeCheckOutSessionRepo(ctx, params, riderID, session.ID, tempRideID)
	if err != nil {
		return nil, err
	}

	return &model.CheckoutSessionResult{
		CheckoutURL:     session.URL,
		StripeSessionID: session.ID,
		TemporaryRideID: tempRideID,
	}, nil
}

func (p *PaymentService) PaymentSuccess(ctx context.Context, params model.PaymentSuccessParams, user_id string) error {

	p_id, err := p.paymentRepo.PaymentSuccessRepo(ctx, params, user_id)

	if err != nil {
		fmt.Println("error payment service :: ", err)
		err = p.sqsQueue.PublishRideRequest(ctx, repository.SQSData{
			Status:     400,
			ServerType: "PAYMENT_SERVER",
			Message:    err.Error(),
			TempRideID: params.TempRideId,
			RiderID:    user_id,
			Aud:        "MAIN_SERVER",
		})

		if err != nil {
			fmt.Println("sendig data to main server from payment on payment success", err)
		}

	}

	err = p.sqsQueue.PublishRideRequest(ctx, repository.SQSData{
		Status:     200,
		ServerType: "PAYMENT_SERVER",
		Message:    "Creating ride",
		TempRideID: params.TempRideId,
		RiderID:    user_id,
		Aud:        "TRIP_SERVER",

		PaymentID: p_id,

		Pickup: repository.Location{
			Address:   params.Pickup.Address,
			Latitude:  params.Pickup.Latitude,
			Longitude: params.Pickup.Longitude,
		},
		Dropoff: repository.Location{
			Address:   params.Dropoff.Address,
			Latitude:  params.Dropoff.Latitude,
			Longitude: params.Dropoff.Longitude,
		},
		Fare: repository.Fare{
			AmountInCents: params.Fare.AmountInCents,
			Currency:      params.Fare.Currency,
		},
		RideDetails: repository.RideDetails{
			DistanceMeters:  params.Ride.DistanceMeters,
			DurationSeconds: params.Ride.DurationSeconds,
			ServiceType:     params.ServiceType,
		},
	})

	if err != nil {
		fmt.Println("sendig data to trip server from payment on payment success", err)
	}
	return nil

}

func (p *PaymentService) ReadSqsStream(ctx context.Context) {
	fmt.Println("starting sqs")
	for {

		messages, err := p.sqsQueue.ReceiveMessages(ctx)

		if err != nil {
			fmt.Printf("error start sqs %s", err.Error())
			continue
		}

		for _, message := range messages {
			if message.Body == nil {
				continue
			}

			var data model.SQSTripCancelRequestFromCancelTrip

			err := json.Unmarshal(
				[]byte(*message.Body),
				&data,
			)

			fmt.Println(data)

			if err != nil {
				fmt.Printf("invalid SQS message: %s", err.Error())
				continue
			}

			if data.Aud != "PAYMENT_SERVER" {
				continue
			}

			p_id := data.PaymentId

			err = p.ProcessRefundWithPaymentIntentId(p_id)

			fmt.Println("error issuing refund ::", err)

			err = p.sqsQueue.DeleteMessage(ctx, *message.ReceiptHandle)

			if err != nil {
				fmt.Println("error deleting message will try again: ", err)
			}

		}
	}
}

func (p *PaymentService) ProcessRefundWithPaymentIntentId(paymentid string) error {

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	paymentIntentId, err := p.paymentRepo.GetPaymentIntentIdWithPaymentID(paymentid)

	if err != nil {
		fmt.Println("Unable to get payment intent :: ", err)
	}

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentId),
	}

	_, err = refund.New(params)
	if err != nil {
		fmt.Println("failed to create refund: %w", err)
		return fmt.Errorf("failed to create refund: %w", err)
	}

	err = p.paymentRepo.UpdateStatusInPaymentDB(paymentIntentId, "refunded")

	if err != nil {
		fmt.Println("failed to update payment refund in db :: %w", err)
	}

	return nil
}

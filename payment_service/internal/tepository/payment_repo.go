package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"payment_service/internal/model"
	"payment_service/proto/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *PaymentRepo {
	return &PaymentRepo{
		db: db,
	}
}

func (p *PaymentRepo) CreateStripeCheckOutSessionRepo(ctx context.Context, params model.CreateCheckoutParams, riderID, stripeSessionID string, tempRideID string) (string, error) {

	serviceType := ""

	if params.ServiceType == pb.ServiceType_SERVICE_TYPE_ECONOMY.String() {
		serviceType = "economy"
	} else if params.ServiceType == pb.ServiceType_SERVICE_TYPE_COMFORT.String() {
		serviceType = "confort"
	} else if params.ServiceType == pb.ServiceType_SERVICE_TYPE_XL.String() {
		serviceType = "XL"
	} else {
		serviceType = "unspecified"
	}

	var paymentID string

	err := p.db.QueryRowContext(ctx, `
		INSERT INTO payment
			(temp_ride_id, stripe_session_id, rider_id, service_type,
			 pickup_latitude, pickup_longitude, pickup_address,
			 dropoff_latitude, dropoff_longitude, dropoff_address,
			 fare_amount_in_cents, fare_currency, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'pending', now())
		RETURNING id`,
		tempRideID, stripeSessionID, riderID, serviceType,
		params.Pickup.Latitude, params.Pickup.Longitude, params.Pickup.Address,
		params.Dropoff.Latitude, params.Dropoff.Longitude, params.Dropoff.Address,
		params.Fare.AmountInCents, params.Fare.Currency,
	).Scan(&paymentID)

	if err != nil {
		log.Printf("failed to insert checkout session: %v", err)
		return "", status.Error(codes.Internal, "failed to persist checkout session")
	}

	return paymentID, nil
}

func (p *PaymentRepo) PaymentSuccessRepo(
	ctx context.Context,
	params model.PaymentSuccessParams,
	userID string,
) (string, error) {

	query := `
		UPDATE payment
		SET status = 'success',
		    updated_at = NOW(),
			stripe_payment_intent_id = $3
		WHERE temp_ride_id = $1
		  AND rider_id = $2
		RETURNING id
	`

	var paymentID string

	err := p.db.QueryRowContext(
		ctx,
		query,
		params.TempRideId,
		userID,
		params.PaymentIntentId,
	).Scan(&paymentID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(
				codes.NotFound,
				"Unable to find payment",
			)
		}

		return "", status.Error(
			codes.Internal,
			"Internal server error",
		)
	}

	return paymentID, nil
}

func (t *PaymentRepo) GetPaymentIntentIdWithPaymentID(payment_id string) (string, error) {
	query := `
		SELECT stripe_payment_intent_id
		FROM payment
		WHERE id = $1
	`

	var payment_intent_id string

	err := t.db.QueryRow(query, payment_id).Scan(&payment_intent_id)
	if err != nil {
		return "", err
	}

	return payment_intent_id, nil
}

func (t *PaymentRepo) UpdateStatusInPaymentDB(paymentID string, status string) error {
	query := `
		UPDATE payment
		SET
			status = $2,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := t.db.Exec(query, paymentID, status)
	if err != nil {
		return err
	}

	return nil
}

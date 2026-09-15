package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"trip_service/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TripRepo struct {
	db *sql.DB
}

func NewTripRepo(db *sql.DB) *TripRepo {
	return &TripRepo{
		db: db,
	}
}

func (r *TripRepo) CreateTrip(
	ctx context.Context,
	params model.SQSDataForCreatingTrip,
	driverAmount int64,
) error {
	query := `
		INSERT INTO trip (
			rider_id,
			payment_id,
			status,
			pickup_address,
			pickup_latitude,
			pickup_longitude,
			dropoff_address,
			dropoff_latitude,
			dropoff_longitude,
			fare,
			driver_amount,
			platform_fee,
			tax_amount,
			discount_amount,
			distance_meters,
			duration_seconds,
			trip_id,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, 'requested',
			$3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16,
			NOW(), NOW()
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		params.RiderID,
		params.PaymentID,
		params.Pickup.Address,
		params.Pickup.Latitude,
		params.Pickup.Longitude,
		params.Dropoff.Address,
		params.Dropoff.Latitude,
		params.Dropoff.Longitude,
		float64(params.Fare.AmountInCents)/100,
		float64(driverAmount)/100,
		0,
		0,
		0,
		params.RideDetails.DistanceMeters,
		params.RideDetails.DurationSeconds,
		params.TempRideID,
	)

	if err != nil {
		fmt.Println("error creating trip", err)
		return err
	}

	return nil
}

func (r *TripRepo) AssignDriver(
	ctx context.Context,
	tripID string,
	riderID string,
	driverID string,
) error {
	query := `
		UPDATE trip
		SET
			driver_id = $1,
			status = 'assigned',
			accepted_at = NOW(),
			updated_at = NOW()
		WHERE trip_id = $2
		  AND rider_id = $3
		  AND status = 'requested'
		  AND driver_id IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		driverID,
		tripID,
		riderID,
	)
	if err != nil {
		return status.Error(
			codes.Internal,
			"An unexpected error occurred while assigning the driver",
		)
	}

	rowCount, err := result.RowsAffected()

	if rowCount == 0 || err != nil {
		return status.Error(
			codes.FailedPrecondition,
			"Trip is no longer available",
		)
	}

	return nil
}

func (r *TripRepo) GetTripByDriverID(
	ctx context.Context,
	driver_id string,
) (*model.Trip, error) {

	query := `
		SELECT
			trip_id,
			rider_id,
			driver_id,
			payment_id,
			status,
			pickup_address,
			pickup_latitude,
			pickup_longitude,
			dropoff_address,
			dropoff_latitude,
			dropoff_longitude,
			fare,
			driver_amount,
			platform_fee,
			tax_amount,
			discount_amount,
			distance_meters,
			duration_seconds,
			cancellation_reason,
			cancelled_by,
			created_at,
			accepted_at,
			picked_up_at,
			completed_at,
			cancelled_at,
			updated_at
			FROM trip
	 		WHERE driver_id = $1
      		AND status IN ('assigned', 'picked')
      		LIMIT 1`
	var trip model.Trip

	err := r.db.QueryRowContext(ctx, query, driver_id).Scan(
		&trip.TripID,
		&trip.RiderID,
		&trip.DriverID,
		&trip.PaymentID,
		&trip.Status,
		&trip.PickupAddress,
		&trip.PickupLatitude,
		&trip.PickupLongitude,
		&trip.DropoffAddress,
		&trip.DropoffLatitude,
		&trip.DropoffLongitude,
		&trip.Fare,
		&trip.DriverAmount,
		&trip.PlatformFee,
		&trip.TaxAmount,
		&trip.DiscountAmount,
		&trip.DistanceMeters,
		&trip.DurationSeconds,
		&trip.CancellationReason,
		&trip.CancelledBy,
		&trip.CreatedAt,
		&trip.AcceptedAt,
		&trip.PickedUpAt,
		&trip.CompletedAt,
		&trip.CancelledAt,
		&trip.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get trip by id: %w", err)
	}

	return &trip, nil
}

func (r *TripRepo) GetTripByRiderId(
	ctx context.Context,
	rider_id string,
) (*model.Trip, error) {

	query := `
		SELECT
			trip_id,
			rider_id,
			driver_id,
			payment_id,
			status,
			pickup_address,
			pickup_latitude,
			pickup_longitude,
			dropoff_address,
			dropoff_latitude,
			dropoff_longitude,
			fare,
			driver_amount,
			platform_fee,
			tax_amount,
			discount_amount,
			distance_meters,
			duration_seconds,
			cancellation_reason,
			cancelled_by,
			created_at,
			accepted_at,
			picked_up_at,
			completed_at,
			cancelled_at,
			updated_at
			FROM trip
	 		WHERE rider_id = $1
      		AND status IN ('assigned', 'picked')
      		LIMIT 1`
	var trip model.Trip

	err := r.db.QueryRow(query, rider_id).Scan(
		&trip.TripID,
		&trip.RiderID,
		&trip.DriverID,
		&trip.PaymentID,
		&trip.Status,
		&trip.PickupAddress,
		&trip.PickupLatitude,
		&trip.PickupLongitude,
		&trip.DropoffAddress,
		&trip.DropoffLatitude,
		&trip.DropoffLongitude,
		&trip.Fare,
		&trip.DriverAmount,
		&trip.PlatformFee,
		&trip.TaxAmount,
		&trip.DiscountAmount,
		&trip.DistanceMeters,
		&trip.DurationSeconds,
		&trip.CancellationReason,
		&trip.CancelledBy,
		&trip.CreatedAt,
		&trip.AcceptedAt,
		&trip.PickedUpAt,
		&trip.CompletedAt,
		&trip.CancelledAt,
		&trip.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get trip by id: %w", err)
	}

	return &trip, nil
}

func (r *TripRepo) RiderPickedUpStatusUpdate(ctx context.Context, tripID string, driverId string) error {
	fmt.Println(tripID)
	query := `
		UPDATE trip
		SET
			status = 'picked',
			picked_up_at = NOW(),
			updated_at = NOW()
		WHERE trip_id = $1
		AND status = 'assigned'
		AND driver_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, tripID, driverId)
	if err != nil {
		return status.Error(
			codes.Internal,
			"failed to update trip status",
		)
	}
	fmt.Println(err)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return status.Error(
			codes.Internal,
			"failed to update trip status",
		)
	}

	if rowsAffected == 0 {
		return status.Error(
			codes.FailedPrecondition,
			"trip is no longer available",
		)
	}

	return nil
}

func (r *TripRepo) CompleteTrip(ctx context.Context, tripId string, driverId string) error {

	query := `
		UPDATE trip
		SET
			status = 'completed',
			completed_at = NOW(),
			updated_at = NOW()
		WHERE
			trip_id = $1
		AND
			driver_id = $2
		AND
			status = 'picked'
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		tripId,
		driverId,
	)
	fmt.Println(err)
	if err != nil {
		return status.Error(
			codes.Internal,
			"failed to complete trip",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return status.Error(
			codes.Internal,
			"failed to complete trip",
		)
	}

	if rowsAffected == 0 {
		return status.Error(
			codes.FailedPrecondition,
			"trip cannot be completed",
		)
	}

	return nil
}
	
	func (r *TripRepo) CancelTrip(
		ctx context.Context,
		tripID string,
		reason string,
		userID string,
		isDriver bool,
	) (string, string, string, error) {
	
		var query string
	
		if isDriver {
			query = `
				UPDATE trip
				SET
					status = 'cancelled',
					cancellation_reason = $2,
					cancelled_at = NOW(),
					updated_at = NOW()
				WHERE trip_id = $1
				  AND status IN ('assigned', 'picked')
				  AND driver_id = $3
				RETURNING payment_id, driver_id, rider_id
			`
		} else {
			query = `
				UPDATE trip
				SET
					status = 'cancelled',
					cancellation_reason = $2,
					cancelled_at = NOW(),
					updated_at = NOW()
				WHERE trip_id = $1
				  AND status IN ('assigned', 'picked', 'searching' , 'requested')
				  AND rider_id = $3
				RETURNING payment_id, driver_id, rider_id
			`
		}
	
		var paymentID string
		var driverID sql.NullString
		var riderID string
	
		err := r.db.QueryRowContext(
			ctx,
			query,
			tripID,
			reason,
			userID,
		).Scan(
			&paymentID,
			&driverID,
			&riderID,
		)
	
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", "", "", status.Error(
					codes.FailedPrecondition,
					"No trip found",
				)
			}
	
			return "", "", "", status.Error(
				codes.Internal,
				"failed to cancel trip",
			)
		}
	
		return paymentID, driverID.String, riderID, nil
	}

func (r *TripRepo) TotalEarningsDriver(
	ctx context.Context,
	driverId string,
) (float32, int32, error) {

	query := `
		SELECT
			COALESCE(SUM(driver_amount), 0),
			COUNT(*)
		FROM trip
		WHERE driver_id = $1
		  AND status = 'completed'
		  AND completed_at >= CURRENT_DATE
		  AND completed_at < CURRENT_DATE + INTERVAL '1 day'
	`

	var totalEarning float32
	var totalTrips int32

	err := r.db.QueryRowContext(
		ctx,
		query,
		driverId,
	).Scan(
		&totalEarning,
		&totalTrips,
	)

	if err != nil {
		return 0, 0, status.Error(
			codes.Internal,
			"failed to get today's earnings",
		)
	}

	return totalEarning, totalTrips, nil
}

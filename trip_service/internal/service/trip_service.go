package service

import (
	"context"
	"encoding/json"
	"fmt"
	"trip_service/internal/model"
	repository "trip_service/internal/tepository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TripService struct {
	sqsRepo  *repository.SQSQueue
	tripRepo *repository.TripRepo
}

func NewTripService(sqsRepo *repository.SQSQueue, tripRepo *repository.TripRepo) *TripService {
	return &TripService{
		sqsRepo:  sqsRepo,
		tripRepo: tripRepo,
	}
}

func (t *TripService) StartListeningSQS(ctx context.Context) {
	for {

		messages, err := t.sqsRepo.ReceiveMessages(ctx)

		if err != nil {
			fmt.Printf("error start sqs %s", err.Error())
			continue
		}

		for _, message := range messages {
			if message.Body == nil {
				continue
			}

			var data model.SQSDataForCreatingTrip

			err := json.Unmarshal(
				[]byte(*message.Body),
				&data,
			)

			if err != nil {
				fmt.Printf("invalid SQS message: %s", err.Error())
				continue
			}

			if data.Aud != "TRIP_SERVER" {
				continue
			}

			driverAmount := data.Fare.AmountInCents * 45 / 100

			err = t.tripRepo.CreateTrip(ctx, data, driverAmount)

			if err != nil {
				t.sqsRepo.SendMessage(ctx, model.SQSDataToMainServer{
					Status:     400,
					ServerType: "TRIP_SERVER",
					Message:    err.Error(),
					TripID:     data.TempRideID,
					RiderID:    data.RiderID,
					Aud:        "MAIN_SERVER",
				})

			}

			//sending data through sqs to main_server so we san serve trip to driver and hopefully they will accept the ride
			err = t.sqsRepo.SendMessage(ctx, model.SQSDataToMainServer{
				Status:     200,
				Aud:        "MAIN_SERVER",
				For:        "TRIP_CREATED",
				Message:    "Trip created successfully",
				TripID:     data.TempRideID, // this in here has become now permanent trip id after above db opreation
				PaymentID:  data.PaymentID,
				RiderID:    data.RiderID,
				DriverFare: driverAmount,
				ServerType: "TRIP_SERVER",
				RideDetails: model.RideDetails{
					DistanceMeters:  data.RideDetails.DistanceMeters,
					DurationSeconds: data.RideDetails.DurationSeconds,
					ServiceType:     data.ServiceType,
				},
				Pickup: model.Location{
					Address:   data.Pickup.Address,
					Latitude:  data.Pickup.Latitude,
					Longitude: data.Pickup.Longitude,
				},
				Dropoff: model.Location{
					Address:   data.Dropoff.Address,
					Latitude:  data.Dropoff.Latitude,
					Longitude: data.Dropoff.Longitude,
				},
			})

			if err != nil {
				fmt.Println("error sending to main server from trip message will try again: ", err)
			}

			err = t.sqsRepo.DeleteMessage(ctx, *message.ReceiptHandle)

			if err != nil {
				fmt.Println("error deleting message will try again: ", err)
			}
		}

	}
}

func (t *TripService) AcceptDriverTripService(ctx context.Context, data model.AcceptDriverTripParams) error {

	err := t.tripRepo.AssignDriver(ctx, data.Trip_id, data.Rider_id, data.Driver_id)

	if err != nil {
		return err
	}

	return nil
}

func (t *TripService) GetTripDetailsWithDriverID(ctx context.Context, driver_id string) (*model.Trip, error) {
	trip, err := t.tripRepo.GetTripByDriverID(ctx, driver_id)

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return trip, nil
}

func (t *TripService) GetTripDetailsWithRiderId(ctx context.Context, rider_id string) (*model.Trip, error) {
	trip, err := t.tripRepo.GetTripByRiderId(ctx, rider_id)

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return trip, nil
}

func (t *TripService) RidePickedUpStatusUpdate(ctx context.Context, trip_id string, driverId string) error {

	err := t.tripRepo.RiderPickedUpStatusUpdate(ctx, trip_id, driverId)

	if err != nil {
		return err
	}

	// t.sqsRepo.SendMessage(ctx, )
	
	return nil
}

func (t *TripService) TripCompletedService(ctx context.Context, tripId string, driverId string, riderId string) error {
	err := t.tripRepo.CompleteTrip(ctx, tripId, driverId)

	if err != nil {
		return err
	}

	data := model.SQSTripCompletedRequestToMain{
		ServerType: "TRIP_SERVER",
		Aud:        "MAIN_SERVER",
		Message:    "Trip completed, notify rider and driver",
		RiderID:    riderId,
		DriverId:   driverId,
		For:        "TRIP_COMPLETE",
		TripId:     tripId,
	}

	err = t.sqsRepo.SendMessage(ctx, data)

	if err != nil {
		fmt.Println("error sending data to main server by sqs on trip complete:: ", err)
	}

	return nil
}

func (t *TripService) CancelTripService(ctx context.Context, trip_id string, isDriver bool, message string, userID string) error {
	payment_id, driver_id, rider_id, err := t.tripRepo.CancelTrip(ctx, trip_id, message, userID, isDriver)

	if err != nil {
		return err
	}

	err = t.sqsRepo.SendMessage(ctx, &model.SQSTripCancelRequestToPayment{
		PaymentId:  payment_id,
		ServerType: "TRIP_SERVER",
		Aud:        "PAYMENT_SERVER",
		Message:    "Process refund",
		RiderID:    rider_id,
		DriverId:   driver_id,
	})

	err = t.sqsRepo.SendMessage(ctx, &model.SQSTripCancelRequestToMain{
		PaymentId:           payment_id,
		ServerType:          "TRIP_SERVER",
		Aud:                 "MAIN_SERVER",
		For:                 "CANCEL_TRIP",
		Message:             "Notify rider and driver that trip is cancelled",
		RiderID:             rider_id,
		DriverId:            driver_id,
		IsCancelledByDriver: isDriver,
		TripId:              trip_id,
	})

	if err != nil {
		fmt.Println(err)
		status.Error(codes.Internal, "Unable to process refund! please call custmore service!")
	}

	return nil
}

func (t *TripService) TotalEarningsTodayService(ctx context.Context, driver_id string) (float32, int32, error) {
	earnings, trips, err := t.tripRepo.TotalEarningsDriver(ctx, driver_id)


	if err != nil {
		return 0, 0, err
	}

	return earnings, (trips), nil
}

package server

import (
	"context"
	"trip_service/internal/middleware"
	"trip_service/internal/model"
	"trip_service/internal/service"
	"trip_service/proto/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TripServer struct {
	pb.UnimplementedTripServiceServer
	tripService *service.TripService
}

func NewTripServer(tripService *service.TripService) *TripServer {
	return &TripServer{
		tripService: tripService,
	}
}

func (t *TripServer) AcceptTripDriver(ctx context.Context, req *pb.AcceptTripDriverRequest) (*pb.AcceptTripDriverResponse, error) {
	user_id, ok := ctx.Value(middleware.Driver_id).(string)

	if !ok || user_id == "" {
		return nil, status.Error(codes.PermissionDenied, "rider id is missing from request context")
	}

	if req.GetRiderId() == "" || req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "All the fields are required")
	}

	err := t.tripService.AcceptDriverTripService(ctx, model.AcceptDriverTripParams{
		Rider_id:  req.GetRiderId(),
		Driver_id: user_id,
		Trip_id:   req.GetTripId(),
	})

	if err != nil {
		return nil, err
	}

	return &pb.AcceptTripDriverResponse{
		TripId:  req.GetTripId(),
		Status:  "assigned",
		Message: "Driver successfully assigned to the trip",
	}, nil

}

func (t *TripServer) GetActiveTrip(ctx context.Context, req *pb.GetActiveTripRequest) (*pb.GetActiveTripResponse, error) {
	driverID, ok := ctx.Value(middleware.Driver_id).(string)
	if !ok || driverID == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	trip, err := t.tripService.GetTripDetailsWithDriverID(ctx, driverID)

	if err != nil {
		return nil, err
	}

	if trip == nil {
		return &pb.GetActiveTripResponse{}, nil
	}

	resp := &pb.GetActiveTripResponse{
		Pickup: &pb.GeoLocation{
			Latitude:  trip.PickupLatitude,
			Longitude: trip.PickupLongitude,
			Address:   trip.PickupAddress,
		},
		Dropoff: &pb.GeoLocation{
			Latitude:  trip.DropoffLatitude,
			Longitude: trip.DropoffLongitude,
			Address:   trip.DropoffAddress,
		},
		RideDetails: &pb.RideDetails{
			DurationSeconds: trip.DurationSeconds,
			DistanceMetrs:   trip.DistanceMeters,
			Status:          trip.Status,
		},

		RiderId:    trip.RiderID,
		DriverFare: trip.DriverAmount,
		TripId:     trip.TripID,
	}

	return resp, nil
}

func (t *TripServer) RiderPickedUp(ctx context.Context, req *pb.RiderPickedUpRequest) (*pb.RiderPickedUpResponse, error) {
	driverID, ok := ctx.Value(middleware.Driver_id).(string)
	if !ok || driverID == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	if req.GetTripId() == "" {
		status.Error(codes.InvalidArgument, "trip id is required")
	}

	err := t.tripService.RidePickedUpStatusUpdate(ctx, req.GetTripId(), driverID)

	if err != nil {
		return nil, err
	}

	return &pb.RiderPickedUpResponse{
		Message:   "Rider picked up",
		TripId:    req.GetTripId(),
		IsSuccess: true,
	}, nil
}

func (t *TripServer) TripCompleted(ctx context.Context, req *pb.TripCompletedRequest) (*pb.TripCompletedResponse, error) {
	driverID, ok := ctx.Value(middleware.Driver_id).(string)
	if !ok || driverID == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip id is required")
	}

	if req.GetRiderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rider id is required")
	}

	err := t.tripService.TripCompletedService(ctx, req.GetTripId(), driverID, req.GetRiderId())

	if err != nil {
		return nil, err
	}

	return &pb.TripCompletedResponse{
		Message:   "Trip completed",
		TripId:    req.GetTripId(),
		IsSuccess: true,
	}, nil

}

func (t *TripServer) GetActiveTripRider(ctx context.Context, req *pb.GetActiveTripRiderRequest) (*pb.GetActiveTripRiderResponse, error) {

	RiderID, ok := ctx.Value(middleware.Rider_id).(string)
	if !ok || RiderID == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	trip, err := t.tripService.GetTripDetailsWithRiderId(ctx, RiderID)

	if err != nil {
		return nil, err
	}

	if trip == nil {
		return &pb.GetActiveTripRiderResponse{}, nil
	}
	
	resp := &pb.GetActiveTripRiderResponse{
		Pickup: &pb.GeoLocation{
			Latitude:  trip.PickupLatitude,
			Longitude: trip.PickupLongitude,
			Address:   trip.PickupAddress,
		},
		Dropoff: &pb.GeoLocation{
			Latitude:  trip.DropoffLatitude,
			Longitude: trip.DropoffLongitude,
			Address:   trip.DropoffAddress,
		},
		RideDetails: &pb.RideDetails{
			DurationSeconds: trip.DurationSeconds,
			DistanceMetrs:   trip.DistanceMeters,
			Status:          trip.Status,
		},

		RiderId:    trip.RiderID,
		DriverFare: trip.DriverAmount,
		TripId:     trip.TripID,
	}

	return resp, nil
}

func (t *TripServer) CancelTrip(ctx context.Context, req *pb.CancelTripRequest) (*pb.CancelTripResponse, error) {
	riderID, _ := ctx.Value(middleware.Rider_id).(string)
	driverID, _ := ctx.Value(middleware.Driver_id).(string)

	if riderID == "" && driverID == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Trip id is requried")
	}

	var userID string
	var isDriver bool

	if riderID != "" {
		userID = riderID
		isDriver = false
	} else {
		userID = driverID
		isDriver = true
	}

	err := t.tripService.CancelTripService(ctx, req.GetTripId(), isDriver, req.GetReason(), userID)

	if err != nil {
		return nil, err
	}

	return &pb.CancelTripResponse{
		Message:   "Success",
		TripId:    req.TripId,
		IsSuccess: true,
	}, nil
}

func (t *TripServer) TotalEarningToday(
	ctx context.Context,
	req *pb.TotalEarningTodayRequest,
) (*pb.TotalEarningTodayResponse, error) {
	driverID, ok := ctx.Value(middleware.Driver_id).(string)
	if !ok || driverID == "" {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	earnings, trips, err := t.tripService.TotalEarningsTodayService(ctx, driverID)

	if err != nil {
		return nil, err
	}

	return &pb.TotalEarningTodayResponse{
		TotalEarningToday: earnings,
		TotalTripsToday:   trips,
	}, nil
}

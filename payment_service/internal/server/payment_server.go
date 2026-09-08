package server

import (
	"context"
	"fmt"
	errors "payment_service/internal/error"
	"payment_service/internal/middleware"
	"payment_service/internal/model"
	"payment_service/internal/service"
	"payment_service/proto/pb"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentService service.PaymentService
}

func NewPaymentServer(paymentService service.PaymentService) *PaymentServer {
	return &PaymentServer{
		paymentService: paymentService,
	}
}

func (p *PaymentServer) CreateCheckoutSession(ctx context.Context, req *pb.CreateCheckoutSessionRequest) (*pb.CreateCheckoutSessionResponse, error) {
	user_id, ok := ctx.Value(middleware.Rider_id).(string)

	if !ok || user_id == "" {
		return nil, status.Error(codes.PermissionDenied, "rider id is missing from request context")
	}
	validation := make(map[string]string)

	if req.GetServiceType() == pb.ServiceType_SERVICE_TYPE_UNSPECIFIED {
		validation["service_type"] = "service_type must be specified (ECONOMY, COMFORT, or XL)"
	}

	pickup := req.GetPickup()
	if pickup == nil {
		validation["pickup"] = "pickup location details are required"
	} else {
		if strings.TrimSpace(pickup.GetAddress()) == "" {
			validation["pickup_address"] = "pickup address is required"
		}
		if pickup.GetLatitude() == 0 {
			validation["pickup_latitude"] = "pickup latitude is required"
		}
		if pickup.GetLongitude() == 0 {
			validation["pickup_longitude"] = "pickup longitude is required"
		}
	}

	dropoff := req.GetDropoff()
	if dropoff == nil {
		validation["dropoff"] = "dropoff location details are required"
	} else {
		if strings.TrimSpace(dropoff.GetAddress()) == "" {
			validation["dropoff_address"] = "dropoff address is required"
		}
		if dropoff.GetLatitude() == 0 {
			validation["dropoff_latitude"] = "dropoff latitude is required"
		}
		if dropoff.GetLongitude() == 0 {
			validation["dropoff_longitude"] = "dropoff longitude is required"
		}
	}

	fare := req.GetFare()
	if fare == nil {
		validation["fare"] = "fare details are required"
	} else {
		if fare.GetAmountInCents() <= 0 {
			validation["fare_amount"] = "fare amount must be greater than zero"
		}
		if strings.TrimSpace(fare.GetCurrency()) == "" {
			validation["fare_currency"] = "fare currency is required"
		}
	}

	ride_details := req.GetRideDetails()

	if ride_details == nil {
		validation["ride_details"] = "ride_details details are required"
	} else {
		if ride_details.GetDurationSeconds() <= 0 {
			validation["duration"] = "duration_seconds must be greater than zero"
		}
		if ride_details.GetDistanceMetrs() <= 0 {
			validation["duration"] = "distance must be greater than zero"
		}
	}

	if len(validation) > 0 {
		return nil, errors.CreateError(validation)
	}

	params := model.CreateCheckoutParams{
		ServiceType: mapServiceType(req.GetServiceType()),

		Pickup: model.GeoLocation{
			Latitude:  req.GetPickup().GetLatitude(),
			Longitude: req.GetPickup().GetLongitude(),
			Address:   req.GetPickup().GetAddress(),
		},
		Dropoff: model.GeoLocation{
			Latitude:  req.GetDropoff().GetLatitude(),
			Longitude: req.GetDropoff().GetLongitude(),
			Address:   req.GetDropoff().GetAddress(),
		},
		Fare: model.FareDetails{
			AmountInCents: req.GetFare().GetAmountInCents(),
			Currency:      req.GetFare().GetCurrency(),
		},

		Ride: model.RideDetails{
			DurationSeconds: req.GetRideDetails().GetDurationSeconds(),
			DistanceMeters:  req.GetRideDetails().GetDistanceMetrs(),
		},
	}

	res, err := p.paymentService.CreateCheckoutSession(ctx, params, user_id)

	if err != nil {
		fmt.Println("creating checkout session inside server", err)
	}

	if err != nil {
		return &pb.CreateCheckoutSessionResponse{}, err
	}

	return &pb.CreateCheckoutSessionResponse{
		CheckoutUrl:     res.CheckoutURL,
		StripeSessionId: res.StripeSessionID,
		TemporaryRideId: res.TemporaryRideID,
	}, nil

}

func mapServiceType(st pb.ServiceType) string {
	switch st {
	case pb.ServiceType_SERVICE_TYPE_ECONOMY:
		return "ECONOMY"
	case pb.ServiceType_SERVICE_TYPE_COMFORT:
		return "COMFORT"
	case pb.ServiceType_SERVICE_TYPE_XL:
		return "XL"
	default:
		return "UNSPECIFIED"
	}
}

func (p *PaymentServer) PaymentSuccess(ctx context.Context, req *pb.PaymentSuccessRequest) (*pb.PaymentSuccessResponse, error) {

	user_id, ok := ctx.Value(middleware.Rider_id).(string)
	if !ok || user_id == "" {
		return nil, status.Error(codes.PermissionDenied, "rider id is missing from request context")
	}
	validation := make(map[string]string)

	pickup := req.GetPickup()
	if pickup == nil {
		validation["pickup"] = "pickup location details are required"
	} else {
		if strings.TrimSpace(pickup.GetAddress()) == "" {
			validation["pickup_address"] = "pickup address is required"
		}
		if pickup.GetLatitude() == 0 {
			validation["pickup_latitude"] = "pickup latitude is required"
		}
		if pickup.GetLongitude() == 0 {
			validation["pickup_longitude"] = "pickup longitude is required"
		}
	}

	dropoff := req.GetDropoff()
	if dropoff == nil {
		validation["dropoff"] = "dropoff location details are required"
	} else {
		if strings.TrimSpace(dropoff.GetAddress()) == "" {
			validation["dropoff_address"] = "dropoff address is required"
		}
		if dropoff.GetLatitude() == 0 {
			validation["dropoff_latitude"] = "dropoff latitude is required"
		}
		if dropoff.GetLongitude() == 0 {
			validation["dropoff_longitude"] = "dropoff longitude is required"
		}
	}

	fare := req.GetFare()
	if fare == nil {
		validation["fare"] = "fare details are required"
	} else {
		if fare.GetAmountInCents() <= 0 {
			validation["fare_amount"] = "fare amount must be greater than zero"
		}
		if strings.TrimSpace(fare.GetCurrency()) == "" {
			validation["fare_currency"] = "fare currency is required"
		}
	}

	ride_details := req.GetRideDetails()

	if ride_details == nil {
		validation["ride_details"] = "ride_details details are required"
	} else {
		if ride_details.GetDurationSeconds() <= 0 {
			validation["duration"] = "duration_seconds must be greater than zero"
		}
		if ride_details.GetDistanceMetrs() <= 0 {
			validation["duration"] = "distance must be greater than zero"
		}
	}

	temp_ride_id := req.GetTempRideId()

	if strings.TrimSpace(temp_ride_id) == "" {
		validation["temp_ride_id"] = "temp_ride_id is required"
	}

	if len(validation) > 0 {
		return nil, errors.CreateError(validation)
	}

	params := model.PaymentSuccessParams{

		Pickup: model.GeoLocation{
			Latitude:  req.GetPickup().GetLatitude(),
			Longitude: req.GetPickup().GetLongitude(),
			Address:   req.GetPickup().GetAddress(),
		},
		Dropoff: model.GeoLocation{
			Latitude:  req.GetDropoff().GetLatitude(),
			Longitude: req.GetDropoff().GetLongitude(),
			Address:   req.GetDropoff().GetAddress(),
		},
		Fare: model.FareDetails{
			AmountInCents: req.GetFare().GetAmountInCents(),
			Currency:      req.GetFare().GetCurrency(),
		},
		TempRideId: temp_ride_id,
		Ride: model.RideDetails{
			DurationSeconds: req.GetRideDetails().GetDurationSeconds(),
			DistanceMeters:  req.GetRideDetails().GetDistanceMetrs(),
		},
	}

	err := p.paymentService.PaymentSuccess(ctx, params, user_id)

	if err != nil {
		fmt.Println("Inside payment success server", err)
	}
	return &pb.PaymentSuccessResponse{}, nil
}

func (p *PaymentServer) PaymentFailed(ctx context.Context, req *pb.PaymentFailedRequest) (*pb.PaymentFailedResponse, error) {
	return &pb.PaymentFailedResponse{}, nil
}

package payment_handler

import (
	"bombe_main_server/internal/middleware/rider_middleware"
	"bombe_main_server/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"payment_service/proto/pb"
	"strconv"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
)

type PaymentHandler struct {
	PaymentClient pb.PaymentServiceClient
}

func NewPaymentHandler(PaymentClient pb.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{
		PaymentClient: PaymentClient,
	}
}

func (p *PaymentHandler) CreateCheckOutSession(w http.ResponseWriter, r *http.Request) {

	riderId, ok := r.Context().Value(rider_middleware.ClaimsContextKey).(string)

	fmt.Println("rideris", riderId)
	if !ok {
		fmt.Println("error: unauthorized rider context in CreateCheckOutSession")
		utils.RespondWithError(w, "Unauthorized rider context", http.StatusUnauthorized)
		return
	}
	internalToken, err := utils.CreateToken(riderId, utils.RoleRider, 1*time.Minute, "payment-grpc-service")

	fmt.Println("creating token error")
	
	if err != nil {
		fmt.Println("error creating internal token for payment-grpc-service:", err)
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Println("error reading request body in CreateCheckOutSession:", err)
		utils.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	grpcReq := pb.CreateCheckoutSessionRequest{}

	err = protojson.Unmarshal(bodyBytes, &grpcReq)

	if err != nil {
		fmt.Println("error unmarshalling request body in CreateCheckOutSession:", err)
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := p.PaymentClient.CreateCheckoutSession((ctx), &grpcReq)

	if err != nil {
		fmt.Println("error calling payment-grpc-service CreateCheckoutSession:", err)
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	httpRes := map[string]any{
		"url": grpcResp.GetCheckoutUrl(),
	}

	utils.RespondWithSuccess(w, "Success", http.StatusOK, httpRes)
}

func (p *PaymentHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Printf("STRIPE_WEBHOOK_SECRET environment variable not set.\n")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	event, err := webhook.ConstructEvent(
		body,
		r.Header.Get("Stripe-Signature"),
		webhookSecret,
	)
	if err != nil {
		log.Printf("Error verifying Stripe webhook signature: %v", err)
		http.Error(w, "invalid webhook", http.StatusBadRequest)
		return
	}

	_, err = json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Printf("Error marshaling Stripe event: %v", err)
		return
	}

	if event.Type == "checkout.session.completed" {
		var cs stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &cs)
		if err != nil {
			log.Printf("Error unmarshaling checkout session: %v", err)
			http.Error(w, "invalid event data", http.StatusInternalServerError)
			return
		}

		rideData, err := extractRideMetadata(cs.Metadata)

		internalToken, err := utils.CreateToken(rideData.RiderID, utils.RoleRider, 1*time.Minute, "payment-grpc-service")
		if err != nil {
			fmt.Println("error creating internal token for payment-grpc-service:", err)
			utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		md := metadata.Pairs("authorization", "Bearer "+internalToken)
		ctx = metadata.NewOutgoingContext(ctx, md)

		_, err = p.PaymentClient.PaymentSuccess(ctx, &pb.PaymentSuccessRequest{
			Pickup: &pb.GeoLocation{
				Latitude:  rideData.PickupLat,
				Longitude: rideData.PickupLng,
				Address:   rideData.PickupAddress,
			},
			Dropoff: &pb.GeoLocation{
				Latitude:  rideData.DropoffLat,
				Longitude: rideData.DropoffLng,
				Address:   rideData.DropoffAddress,
			},
			Fare: &pb.FareDetails{
				AmountInCents: rideData.Amount,
				Currency:      "usd",
			},
			// ServiceType: pb.ServiceType_SERVICE_TYPE_UNSPECIFIED,
			TempRideId: rideData.TempRideId,
			RiderId:    rideData.RiderID,
			RideDetails: &pb.RideDetails{
				DurationSeconds: rideData.DurationSeconds,
				DistanceMetrs:   rideData.DistanceMeters,
			},
		})

		if err != nil {
			fmt.Println("error from payment success webhook ", err)
		}

	}

	if event.Type == "payment_intent.payment_failed" {
		var pi stripe.PaymentIntent

		err := json.Unmarshal(event.Data.Raw, &pi)
		if err != nil {
			log.Printf("Error unmarshaling payment intent: %v", err)
			http.Error(w, "invalid event data", http.StatusInternalServerError)
			return
		}

		fmt.Println("Payment Intent Metadata:")
		rideData, err := extractRideMetadata(pi.Metadata)
		fmt.Println(rideData)
	}

}

func extractRideMetadata(metadata map[string]string) (RideMetadata, error) {
	pickupLat, err := strconv.ParseFloat(metadata["pickup_lat"], 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid pickup_lat: %w", err)
	}

	pickupLng, err := strconv.ParseFloat(metadata["pickup_lng"], 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid pickup_lng: %w", err)
	}

	dropoffLat, err := strconv.ParseFloat(metadata["dropoff_lat"], 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid dropoff_lat: %w", err)
	}

	dropoffLng, err := strconv.ParseFloat(metadata["dropoff_lng"], 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid dropoff_lng: %w", err)
	}

	amount, err := strconv.ParseInt(metadata["amount"], 10, 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid amount: %w", err)
	}

	distanceMeters, err := strconv.ParseInt(metadata["distance_meters"], 10, 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid distance in meters: %w", err)
	}

	durationSeconds, err := strconv.ParseInt(metadata["duration_seconds"], 10, 64)
	if err != nil {
		return RideMetadata{}, fmt.Errorf("invalid duration in seconds: %w", err)
	}

	return RideMetadata{
		RiderID:         metadata["rider_id"],
		ServiceType:     metadata["service_type"],
		PickupAddress:   metadata["pickup_location"],
		DropoffAddress:  metadata["dropoff_location"],
		PickupLat:       pickupLat,
		PickupLng:       pickupLng,
		DropoffLat:      dropoffLat,
		DropoffLng:      dropoffLng,
		Amount:          amount,
		TempRideId:      metadata["temp_ride_id"],
		DurationSeconds: durationSeconds,
		DistanceMeters:  distanceMeters,
	}, nil
}

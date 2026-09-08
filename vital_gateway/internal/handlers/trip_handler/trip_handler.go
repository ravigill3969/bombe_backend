package trip_handler

import (
	"bombe_main_server/internal/middleware/driver_middleware"
	"bombe_main_server/internal/middleware/rider_middleware"
	"bombe_main_server/internal/utils"
	"fmt"
	"io"
	"net/http"
	"time"
	"trip_service/proto/pb"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type TripHandler struct {
	TripClient pb.TripServiceClient
}

func NewTripHandler(tripClient pb.TripServiceClient) *TripHandler {
	return &TripHandler{
		TripClient: tripClient,
	}
}

func (t *TripHandler) AssignTripToDriver(w http.ResponseWriter, r *http.Request) {
	driverId, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverId, "driver", 1*time.Minute, "trip-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Println("error reading request body in Assigning trip:", err)
		utils.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	grpcReq := pb.AcceptTripDriverRequest{}

	err = protojson.Unmarshal(bodyBytes, &grpcReq)

	if err != nil {
		fmt.Println("error unmarshalling request body in Assigning trip:", err)
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := t.TripClient.AcceptTripDriver((ctx), &grpcReq)

	if err != nil {
		fmt.Println("error calling trip-grpc-service Assigning trip:", err)
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	utils.RespondWithSuccess(w, grpcResp.GetMessage(), http.StatusCreated, "")

}

func (t *TripHandler) GetTripWithDriverId(w http.ResponseWriter, r *http.Request) {
	driverId, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverId, "driver", 1*time.Minute, "trip-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := t.TripClient.GetActiveTrip(ctx, &pb.GetActiveTripRequest{})

	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	jsonBytes, err := protojson.Marshal(grpcResp)
	if err != nil {
		fmt.Printf("error while responding get active trip with driver id: %v\n", err)
		utils.RespondWithError(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(http.StatusOK))
	w.Write(jsonBytes)

}

func (t *TripHandler) UpdateTripStatusToPicked(w http.ResponseWriter, r *http.Request) {
	driverId, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverId, "driver", 1*time.Minute, "trip-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Println("error reading request body in picked trip:", err)
		utils.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	grpcReq := pb.RiderPickedUpRequest{}

	err = protojson.Unmarshal(bodyBytes, &grpcReq)

	if err != nil {
		fmt.Println("error unmarshalling request body in Assigning trip:", err)
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := t.TripClient.RiderPickedUp((ctx), &grpcReq)

	if err != nil {
		fmt.Println("error calling trip-grpc-service picked trip:", err)
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	utils.RespondWithSuccess(w, grpcResp.GetMessage(), http.StatusCreated, "")

}

func (t *TripHandler) MarkTripCompleted(w http.ResponseWriter, r *http.Request) {
	driverId, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverId, "driver", 1*time.Minute, "trip-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		fmt.Println("error reading request body in completed trip:", err)
		utils.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	grpcReq := pb.TripCompletedRequest{}

	err = protojson.Unmarshal(bodyBytes, &grpcReq)

	if err != nil {
		fmt.Println("error unmarshalling request body in completing trip:", err)
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := t.TripClient.TripCompleted((ctx), &grpcReq)

	if err != nil {
		fmt.Println("error calling trip-grpc-service completed trip:", err)
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	utils.RespondWithSuccess(w, grpcResp.GetMessage(), http.StatusCreated, "")

}

func (t *TripHandler) GetTripWithRiderId(w http.ResponseWriter, r *http.Request) {
	riderId, ok := r.Context().Value(rider_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(riderId, "rider", 1*time.Minute, "trip-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := t.TripClient.GetActiveTripRider(ctx, &pb.GetActiveTripRiderRequest{})

	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	jsonBytes, err := protojson.Marshal(grpcResp)
	if err != nil {
		fmt.Printf("error while responding get active trip with rider id: %v\n", err)
		utils.RespondWithError(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(http.StatusOK))
	w.Write(jsonBytes)

}

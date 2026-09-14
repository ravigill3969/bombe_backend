package rider_handler

import (
	"auth_service/proto/pb"
	"bombe_main_server/internal/middleware/driver_middleware"
	"bombe_main_server/internal/middleware/rider_middleware"
	"bombe_main_server/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type AuthHandler struct {
	AuthClient pb.RiderAuthServiceClient
}

func NewAuthHandler(authClient pb.RiderAuthServiceClient) *AuthHandler {
	return &AuthHandler{
		AuthClient: authClient,
	}
}

func (d *AuthHandler) RegisterRider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		utils.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	grpcReq := &pb.RiderRegisterRequest{}

	err = protojson.Unmarshal(bodyBytes, grpcReq)
	if err != nil {
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := d.AuthClient.RegisterRider(r.Context(), grpcReq)

	if err != nil {
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	access_cookie := http.Cookie{
		Name:     "rider_access_token",
		Value:    grpcResp.AccessToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &access_cookie)

	utils.RespondWithSuccess(w, "Logged-in successfully", 200, nil)
}

func (d *AuthHandler) LoginRider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		utils.RespondWithError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	grpcReq := &pb.RiderLoginRequest{}

	err = protojson.Unmarshal(bodyBytes, grpcReq)
	if err != nil {
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := d.AuthClient.LoginRider(r.Context(), grpcReq)

	if err != nil {
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	access_cookie := http.Cookie{
		Name:     "rider_access_token",
		Value:    grpcResp.AccessToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &access_cookie)

	utils.RespondWithSuccess(w, "Logged-in successfully", 200, nil)
}

func (d *AuthHandler) GetRiderInfoo(w http.ResponseWriter, r *http.Request) {

	riderId, ok := r.Context().Value(rider_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized rider context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(riderId, "rider", 1*time.Minute, "auth-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := d.AuthClient.GetRiderInfo(ctx, &pb.GetRiderInfoRequest{})
	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	response := map[string]any{
		"firstname":    grpcResp.GetFirstname(),
		"lastname":     grpcResp.GetLastname(),
		"email":        grpcResp.GetEmail(),
		"phone_number": grpcResp.GetPhoneNumber(),
		"image_url":    grpcResp.GetImageUrl(),
		"id":           grpcResp.GetRiderId(),
		"rating":       grpcResp.GetRating(),
	}
	utils.RespondWithSuccess(w, "validated", http.StatusOK, response)

}

func (d *AuthHandler) LogoutRider(w http.ResponseWriter, r *http.Request) {
	riderId, ok := r.Context().Value(rider_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(riderId, "rider", 1*time.Minute, "auth-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := d.AuthClient.LogoutRider(ctx, &pb.LogoutRiderRequest{})
	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	access_cookie := http.Cookie{
		Name:     "rider_access_token",
		Value:    grpcResp.AccessToken,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &access_cookie)

	utils.RespondWithSuccess(w, "Logged out successfully", http.StatusOK, nil)
}

func (d *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	riderId, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized rider context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(riderId, "rider", 1*time.Minute, "auth-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	var data UpdatePasswordRequest

	err = json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		utils.RespondWithError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	grpcResp, err := d.AuthClient.UpdatePasswordRider(ctx, &pb.UpdatePasswordRiderRequest{
		CurrentPassword:    data.CurrentPassword,
		NewPassword:        data.NewPassword,
		ConfirmNewPassword: data.ConfirmNewPassword,
	})
	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	jsonBytes, err := protojson.Marshal(grpcResp)
	if err != nil {
		fmt.Printf("error while responding updatePassword : %v\n", err)
		utils.RespondWithError(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(http.StatusOK))
	w.Write(jsonBytes)
}

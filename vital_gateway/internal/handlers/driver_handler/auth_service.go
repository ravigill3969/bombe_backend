package driver_handler

import (
	"auth_service/proto/pb"
	"bombe_main_server/internal/middleware/driver_middleware"
	"bombe_main_server/internal/utils"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type AuthHandler struct {
	AuthClient pb.DriverAuthServiceClient
}

func NewAuthHandler(authClient pb.DriverAuthServiceClient) *AuthHandler {
	return &AuthHandler{
		AuthClient: authClient,
	}
}

func (d *AuthHandler) RegisterDriver(w http.ResponseWriter, r *http.Request) {
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

	grpcReq := &pb.DriverRegisterRequest{}

	err = protojson.Unmarshal(bodyBytes, grpcReq)
	if err != nil {
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := d.AuthClient.RegisterDriver(r.Context(), grpcReq)

	if err != nil {
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	jsonBytes, err := protojson.Marshal(grpcResp)
	if err != nil {
		utils.RespondWithError(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(int(grpcResp.StatusCode))
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func (d *AuthHandler) LoginDriver(w http.ResponseWriter, r *http.Request) {
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

	grpcReq := &pb.DriverLoginRequest{}

	err = protojson.Unmarshal(bodyBytes, grpcReq)
	if err != nil {
		utils.RespondWithError(w, "Invalid JSON format for gRPC type: "+err.Error(), http.StatusBadRequest)
		return
	}

	grpcResp, err := d.AuthClient.LoginDriver(r.Context(), grpcReq)

	if err != nil {
		status_code, status := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, status, status_code)
		return
	}

	access_cookie := http.Cookie{
		Name:     "driver_access_token",
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

func (d *AuthHandler) GetDriverInfoo(w http.ResponseWriter, r *http.Request) {

	driverID, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverID, "driver", 1*time.Minute, "auth-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := d.AuthClient.GetDriverInfo(ctx, &pb.GetDriverInfoRequest{})
	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	jsonBytes, err := protojson.Marshal(grpcResp)
	if err != nil {
		fmt.Printf("error while responding getdriverinfo: %v\n", err)
		utils.RespondWithError(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(grpcResp.StatusCode))
	w.Write(jsonBytes)
}

func (d *AuthHandler) LogoutDriver(w http.ResponseWriter, r *http.Request) {
	driverID, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverID, "driver", 1*time.Minute, "auth-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := d.AuthClient.LogoutDriver(ctx, &pb.LogoutDriverRequest{})
	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	access_cookie := http.Cookie{
		Name:     "driver_access_token",
		Value:    grpcResp.AccessToken,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &access_cookie)

	utils.RespondWithSuccess(w, "Logged out successfully", http.StatusOK, nil)
}

func (d *AuthHandler) GetDriverCarInfoo(w http.ResponseWriter, r *http.Request) {

	driverID, ok := r.Context().Value(driver_middleware.ClaimsContextKey).(string)

	if !ok {
		utils.RespondWithError(w, "Unauthorized driver context", http.StatusUnauthorized)
		return
	}

	internalToken, err := utils.CreateToken(driverID, "driver", 1*time.Minute, "auth-grpc-service")
	if err != nil {
		utils.RespondWithError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	md := metadata.Pairs("authorization", "Bearer "+internalToken)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	grpcResp, err := d.AuthClient.GetDriverCarInfo(ctx, &pb.GetDriverCarInfoRequest{})
	if err != nil {
		statusCode, statusMsg := utils.GRPCtoHTTPStatus(err)
		utils.RespondWithError(w, statusMsg, statusCode)
		return
	}

	jsonBytes, err := protojson.Marshal(grpcResp)
	if err != nil {
		fmt.Printf("error while responding getdriverinfo: %v\n", err)
		utils.RespondWithError(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(http.StatusOK))
	w.Write(jsonBytes)
}

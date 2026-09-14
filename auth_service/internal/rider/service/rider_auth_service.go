package rider_service

import (
	"auth_service/internal/middleware"
	rider_repository "auth_service/internal/rider/tepository"
	pb "auth_service/proto/pb"
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RiderAuthService struct {
	pb.UnimplementedRiderAuthServiceServer
	riderAuthRepo *rider_repository.RiderAuthRepo
}

func NewRiderAuthService(repo *rider_repository.RiderAuthRepo) *RiderAuthService {
	return &RiderAuthService{
		riderAuthRepo: repo,
	}
}

func (d *RiderAuthService) RiderRegister(ctx context.Context, req *pb.RiderRegisterRequest) RegisterResponse {
	res := d.riderAuthRepo.RegisterRider(ctx, req)

	if res.Err != nil {
		return RegisterResponse{
			Message:    res.Msg,
			Err:        res.Err,
			StatusCode: int16(res.StatusCode),
		}
	}

	jwt_access_token := os.Getenv("JWT_RIDER_ACCESS_TOKEN_SECRET")

	jwt_refresh_token := os.Getenv("JWT_RIDER_REFRESH_TOKEN_SECRET")

	access_token, err := createToken(res.Rider_id, 7, jwt_access_token)
	if err != nil {
		return RegisterResponse{
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your login"),
			StatusCode: 500,
			Status:     "Internal server error",
		}
	}

	refresh_token, err := createToken(res.Rider_id, 30, jwt_refresh_token)
	if err != nil {
		return RegisterResponse{
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your login"),
			StatusCode: 500,
			Status:     "Internal server error",
		}
	}

	return RegisterResponse{
		Message:       res.Msg,
		Err:           nil,
		Access_token:  access_token,
		Refresh_token: refresh_token,
		StatusCode:    201,
		Status:        "Ok",
	}
}

func (d *RiderAuthService) RiderLogin(ctx context.Context, req *pb.RiderLoginRequest) LoginResponse {
	res := d.riderAuthRepo.LoginRider(ctx, req)

	if res.Err != nil {
		return LoginResponse{
			Message:    res.Msg,
			Err:        res.Err,
			StatusCode: int16(res.StatusCode),
		}
	}

	jwt_access_token := os.Getenv("JWT_RIDER_ACCESS_TOKEN_SECRET")

	jwt_refresh_token := os.Getenv("JWT_RIDER_REFRESH_TOKEN_SECRET")

	access_token, err := createToken(res.Rider_id, 7, jwt_access_token)
	if err != nil {
		return LoginResponse{
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your login"),
			StatusCode: 500,
			Status:     "Internal server error",
		}
	}

	refresh_token, err := createToken(res.Rider_id, 30, jwt_refresh_token)
	if err != nil {
		return LoginResponse{
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your login"),
			StatusCode: 500,
			Status:     "Internal server error",
		}
	}

	return LoginResponse{
		Message:       res.Msg,
		Err:           nil,
		Access_token:  access_token,
		Refresh_token: refresh_token,
		StatusCode:    200,
		Status:        "Ok",
	}
}

func (d *RiderAuthService) GetRiderInfoService(ctx context.Context, req *pb.GetRiderInfoRequest) GetRiderInfoServiceResponse {

	riderId, ok := ctx.Value(middleware.Rider_id).(string)

	if !ok {
		return GetRiderInfoServiceResponse{
			Err:        status.Error(codes.Internal, "Internal server error"),
			Status:     "Internal server error",
			StatusCode: int16(codes.Internal),
			Message:    "Rider_id is required",
		}
	}

	res := d.riderAuthRepo.GetRiderInfoRepo(ctx, riderId)

	if res.Err != nil {
		return GetRiderInfoServiceResponse{
			Message:    res.Msg,
			Err:        res.Err,
			Status:     res.Status,
			StatusCode: int16(res.StatusCode),
		}
	}

	rider_info := res.Rider_info

	return GetRiderInfoServiceResponse{
		Rider_info: GetRiderInfo{
			Rider_id:     riderId,
			Firstname:    rider_info.Firstname,
			Lastname:     rider_info.Lastname,
			Phone_number: rider_info.Phone_number,
			Email:        rider_info.Email,
			ImageURL:     rider_info.ImageURL,
			Rating:       rider_info.Rating,
		},
		Message:    res.Msg,
		Err:        nil,
		Status:     res.Status,
		StatusCode: int16(res.StatusCode),
	}
}

func createToken(rider_id uuid.UUID, day time.Duration, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"rider_id": rider_id,
			"exp":      time.Now().Add(time.Hour * 24 * day * 1).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (d *RiderAuthService) UpdateDriverPasswordService(ctx context.Context, curr_password string, new_password string, rider_id string) error {
	err := d.riderAuthRepo.UpdatePassword(ctx, rider_id, new_password, curr_password)

	if err != nil{
		return  err
	}

	return nil
}
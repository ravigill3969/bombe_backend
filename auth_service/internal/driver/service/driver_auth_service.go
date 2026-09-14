package driver_service

import (
	driver_repository "auth_service/internal/driver/tepository"
	"auth_service/internal/middleware"
	"auth_service/proto/pb"
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DriverAuthService struct {
	pb.UnimplementedDriverAuthServiceServer
	driverAuthRepo *driver_repository.DriverAuthRepo
}

func NewDriverAuthService(repo *driver_repository.DriverAuthRepo) *DriverAuthService {
	return &DriverAuthService{
		driverAuthRepo: repo,
	}
}

func (d *DriverAuthService) DriverRegister(ctx context.Context, req *pb.DriverRegisterRequest) RegisterResponse {
	res := d.driverAuthRepo.RegisterDriver(ctx, req)

	if res.Err != nil {
		return RegisterResponse{
			Message:    res.Msg,
			Err:        res.Err,
			StatusCode: int16(res.StatusCode),
		}
	}

	jwt_access_token := os.Getenv("JWT_DRIVER_ACCESS_TOKEN_SECRET")

	jwt_refresh_token := os.Getenv("JWT_DRIVER_REFRESH_TOKEN_SECRET")

	access_token, err := createToken(res.Driver_id, 7, jwt_access_token)
	if err != nil {
		return RegisterResponse{
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your login"),
			StatusCode: 500,
			Status:     "Internal server error",
		}
	}

	refresh_token, err := createToken(res.Driver_id, 30, jwt_refresh_token)
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

func (d *DriverAuthService) DriverLogin(ctx context.Context, req *pb.DriverLoginRequest) LoginResponse {
	res := d.driverAuthRepo.LoginDriver(ctx, req)

	if res.Err != nil {
		return LoginResponse{
			Message:    res.Msg,
			Err:        res.Err,
			StatusCode: int16(res.StatusCode),
		}
	}

	jwt_access_token := os.Getenv("JWT_DRIVER_ACCESS_TOKEN_SECRET")

	jwt_refresh_token := os.Getenv("JWT_DRIVER_REFRESH_TOKEN_SECRET")

	access_token, err := createToken(res.Driver_id, 7, jwt_access_token)
	if err != nil {
		return LoginResponse{
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your login"),
			StatusCode: 500,
			Status:     "Internal server error",
		}
	}

	refresh_token, err := createToken(res.Driver_id, 30, jwt_refresh_token)
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

func (d *DriverAuthService) GetDriverInfoService(ctx context.Context, req *pb.GetDriverInfoRequest) GetDriverInfoServiceResponse {

	driverID, ok := ctx.Value(middleware.Driver_id).(string)

	if !ok {
		return GetDriverInfoServiceResponse{
			Err:        status.Error(codes.Internal, "Internal server error"),
			Status:     "Internal server error",
			StatusCode: int16(codes.Internal),
			Message:    "Driver_id is required",
		}
	}

	res := d.driverAuthRepo.GetDriverInfoRepo(ctx, driverID)

	if res.Err != nil {
		return GetDriverInfoServiceResponse{
			Err:        res.Err,
			Status:     res.Status,
			StatusCode: int16(res.StatusCode),
			Message:    res.Msg,
		}
	}

	driver_info := res.Driver_info

	return GetDriverInfoServiceResponse{
		Driver_info: GetDriveInfo{
			Driver_id:      driverID,
			Firstname:      driver_info.Firstname,
			Lastname:       driver_info.Lastname,
			Phone_number:   driver_info.Phone_number,
			Email:          driver_info.Email,
			ImageURL:       driver_info.ImageURL,
			License_number: driver_info.License_number,
			Rating:         driver_info.Rating,
		},
		Message:    res.Msg,
		Err:        nil,
		Status:     res.Status,
		StatusCode: int16(res.StatusCode),
	}
}

func (d *DriverAuthService) GetDriverCarInfoService(ctx context.Context, driver_id string) (*pb.GetDriverCarInfoResponse, error) {
	res, err := d.driverAuthRepo.GetDriverCarInfoRepo(driver_id, ctx)

	if err != nil {
		return &pb.GetDriverCarInfoResponse{}, err
	}

	return &pb.GetDriverCarInfoResponse{
		VehicleId:         res.VehicleID.String(),
		Carname:           res.CarName,
		Brand:             res.Brand,
		Model:             res.Model,
		Make:              res.Make,
		Year:              int32(res.Year),
		Coloe:             res.Color,
		CarPlate:          res.CarPlate,
		InsurancePolicyNo: res.InsurancePolicyNo,
		IsVerified:        res.IsVerified,
	}, nil
}

func createToken(driver_id uuid.UUID, day time.Duration, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"driver_id": driver_id,
			"exp":       time.Now().Add(time.Hour * 24 * day * 1).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (d *DriverAuthService) UpdateDriverPasswordService(ctx context.Context, curr_password string, new_password string, driver_id string) error {
	err := d.driverAuthRepo.UpdatePassword(ctx, driver_id, new_password, curr_password)

	if err != nil{
		return  err
	}

	return nil
}

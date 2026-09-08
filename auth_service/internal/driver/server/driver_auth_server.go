package driver_server

import (
	driver_service "auth_service/internal/driver/service"
	"auth_service/internal/middleware"
	pb "auth_service/proto/pb"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DriverAuthServer struct {
	pb.UnimplementedDriverAuthServiceServer
	driverAuthService *driver_service.DriverAuthService
}

func NewDriverAuthServer(service *driver_service.DriverAuthService) *DriverAuthServer {
	return &DriverAuthServer{
		driverAuthService: service,
	}
}

func (s *DriverAuthServer) RegisterDriver(ctx context.Context, req *pb.DriverRegisterRequest) (*pb.DriverAuthResponse, error) {
	res := s.driverAuthService.DriverRegister(ctx, req)

	if res.Err != nil {
		return &pb.DriverAuthResponse{}, res.Err
	}

	return &pb.DriverAuthResponse{
		AccessToken:  res.Access_token,
		RefreshToken: res.Refresh_token,
		StatusCode:   int32(res.StatusCode),
		Message:      res.Message,
		Success:      true,
	}, nil
}

func (s *DriverAuthServer) LoginDriver(ctx context.Context, req *pb.DriverLoginRequest) (*pb.DriverAuthResponse, error) {
	res := s.driverAuthService.DriverLogin(ctx, req)

	if res.Err != nil {
		return &pb.DriverAuthResponse{}, res.Err
	}

	return &pb.DriverAuthResponse{
		AccessToken:  res.Access_token,
		RefreshToken: res.Refresh_token,
		StatusCode:   int32(res.StatusCode),
		Message:      res.Message,
		Success:      true,
	}, nil
}

func (s *DriverAuthServer) GetDriverInfo(ctx context.Context, req *pb.GetDriverInfoRequest) (*pb.GetDriverInfoResponse, error) {
	res := s.driverAuthService.GetDriverInfoService(ctx, req)

	if res.Err != nil {
		return &pb.GetDriverInfoResponse{}, res.Err
	}

	driver_info := res.Driver_info

	return &pb.GetDriverInfoResponse{
		Firstname:   driver_info.Firstname,
		Lastname:    driver_info.Lastname,
		Email:       driver_info.Email,
		PhoneNumber: driver_info.Phone_number,
		LicenseNo:   driver_info.License_number,
		ImageUrl:    driver_info.ImageURL,
		DriverId:    driver_info.Driver_id,
		Rating:      driver_info.Rating,
		StatusCode:  int32(res.StatusCode),
		Message:     res.Message,
		Success:     true,
	}, nil

}

func (s *DriverAuthServer) LogoutDriver(ctx context.Context, req *pb.LogoutDriverRequest) (*pb.LogoutDriverResponse, error) {
	_, ok := ctx.Value(middleware.Driver_id).(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	return &pb.LogoutDriverResponse{
		AccessToken:  "",
		RefreshToken: "",
		Success:      true,
		StatusCode:   200,
		Message:      "Logout success",
	}, nil
}

func (s *DriverAuthServer) GetDriverCarInfo(ctx context.Context, req *pb.GetDriverCarInfoRequest) (*pb.GetDriverCarInfoResponse, error) {
	driver_id, ok := ctx.Value(middleware.Driver_id).(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	return s.driverAuthService.GetDriverCarInfoService(ctx, driver_id)
}

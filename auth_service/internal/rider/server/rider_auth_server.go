package rider_server

import (
	"auth_service/internal/middleware"
	rider_service "auth_service/internal/rider/service"
	pb "auth_service/proto/pb"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RiderAuthServer struct {
	pb.UnimplementedRiderAuthServiceServer
	riderAuthService *rider_service.RiderAuthService
}

func NewRiderAuthServer(service *rider_service.RiderAuthService) *RiderAuthServer {
	return &RiderAuthServer{
		riderAuthService: service,
	}
}

func (r *RiderAuthServer) RegisterRider(ctx context.Context, req *pb.RiderRegisterRequest) (*pb.RiderAuthResponse, error) {
	res := r.riderAuthService.RiderRegister(ctx, req)

	if res.Err != nil {
		return &pb.RiderAuthResponse{}, res.Err
	}

	return &pb.RiderAuthResponse{
		AccessToken:  res.Access_token,
		RefreshToken: res.Refresh_token,
		Success:      true,
		StatusCode:   int32(res.StatusCode),
		Message:      res.Message,
	}, nil
}

func (r *RiderAuthServer) LoginRider(ctx context.Context, req *pb.RiderLoginRequest) (*pb.RiderAuthResponse, error) {
	res := r.riderAuthService.RiderLogin(ctx, req)

	if res.Err != nil {
		return &pb.RiderAuthResponse{}, res.Err
	}

	return &pb.RiderAuthResponse{
		AccessToken:  res.Access_token,
		RefreshToken: res.Refresh_token,
		Success:      true,
		StatusCode:   int32(res.StatusCode),
		Message:      res.Message,
	}, nil
}

func (r *RiderAuthServer) GetRiderInfo(ctx context.Context, req *pb.GetRiderInfoRequest) (*pb.GetRiderInfoResponse, error) {
	res := r.riderAuthService.GetRiderInfoService(ctx, req)

	if res.Err != nil {
		return &pb.GetRiderInfoResponse{}, res.Err
	}

	rider_info := res.Rider_info

	return &pb.GetRiderInfoResponse{
		RiderId:     rider_info.Rider_id,
		Firstname:   rider_info.Firstname,
		Lastname:    rider_info.Lastname,
		Email:       rider_info.Email,
		PhoneNumber: rider_info.Phone_number,
		ImageUrl:    rider_info.ImageURL,
		Rating:      rider_info.Rating,
		Success:     true,
		StatusCode:  int32(res.StatusCode),
		Message:     res.Message,
	}, nil

}

func (r *RiderAuthServer) LogoutRider(ctx context.Context, req *pb.LogoutRiderRequest) (*pb.LogoutRiderResponse, error) {
	_, ok := ctx.Value(middleware.Rider_id).(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	return &pb.LogoutRiderResponse{
		AccessToken:  "",
		RefreshToken: "",
		Success:      true,
		StatusCode:   200,
		Message:      "Logout success",
	}, nil
}

func (s *RiderAuthServer) UpdatePasswordRider(ctx context.Context, req *pb.UpdatePasswordRiderRequest) (*pb.UpdatePasswordRiderResponse, error) {
	driver_id, ok := ctx.Value(middleware.Driver_id).(string)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	if req.GetConfirmNewPassword() != req.GetNewPassword() {
		return nil, status.Error(codes.InvalidArgument, "New password doesnot match")
	}

	err := s.riderAuthService.UpdateDriverPasswordService(ctx, req.GetCurrentPassword(), req.GetNewPassword(), driver_id)

	if err != nil {
		return nil, err
	}

	return &pb.UpdatePasswordRiderResponse{
		Message:   "Password updated successfully",
		IsSuccess: true,
	}, nil

}

package rider_repository

import (
	"auth_service/internal/utils"
	"auth_service/proto/pb"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RiderAuthRepo struct {
	db *sql.DB
}

func NewRiderAuthRepo(db *sql.DB) *RiderAuthRepo {
	return &RiderAuthRepo{
		db: db,
	}
}
func (r *RiderAuthRepo) RegisterRider(ctx context.Context, req *pb.RiderRegisterRequest) RegisterRepResponse {
	if req == nil {
		return RegisterRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Rider registration details are required"),
			Msg:        "Validation failed: RiderRegisterRequest is nil",
			Status:     "Bad Request",
			StatusCode: uint32(codes.InvalidArgument),
		}
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return RegisterRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during registration"),
			Msg:        fmt.Sprintf("Crypto error: failed to hash password: %v", err),
			Status:     "Internal Server Error",
			StatusCode: 500,
		}
	}

	var rider_id uuid.UUID

	query := `INSERT INTO rider_info(
		firstname, lastname, email, password_hash, phone_number, image_url
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING rider_id`

	err = r.db.QueryRowContext(
		ctx,
		query,
		req.Firstname,
		req.Lastname,
		req.Email,
		passwordHash,
		req.PhoneNumber,
		req.ImageUrl,
	).Scan(&rider_id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			clientMsg := "The provided information is already registered"
			if pqErr.Constraint == "rider_info_email_key" {
				clientMsg = "The email address is already registered"
			} else if pqErr.Constraint == "rider_info_phone_number_key" {
				clientMsg = "The phone number is already registered"
			}

			return RegisterRepResponse{
				Rider_id:   uuid.Nil,
				Err:        status.Error(codes.AlreadyExists, clientMsg),
				Msg:        fmt.Sprintf("Conflict: unique constraint violation on %s", pqErr.Constraint),
				Status:     "Conflict",
				StatusCode: uint32(codes.AlreadyExists),
			}
		}

		return RegisterRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during registration"),
			Msg:        fmt.Sprintf("SQL execution error: failed to insert rider_info: %v", err),
			Status:     "InternalError",
			StatusCode: uint32(codes.Internal),
		}
	}

	return RegisterRepResponse{
		Rider_id:   rider_id,
		Err:        nil,
		Msg:        "Success: Rider account registered successfully",
		Status:     "Ok",
		StatusCode: 201,
	}
}

func (r *RiderAuthRepo) LoginRider(ctx context.Context, req *pb.RiderLoginRequest) LoginRepResponse {

	if req.Email == "" {
		return LoginRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Rider email is required"),
			Msg:        "Validation failed: email is nil",
			Status:     "Bad Request",
			StatusCode: 400,
		}
	}

	if req.Password == "" {
		return LoginRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Rider password is required"),
			Msg:        "Validation failed: password is nil",
			Status:     "Bad Request",
			StatusCode: 400,
		}
	}

	query := `SELECT password_hash, rider_id FROM rider_info WHERE email =$1`

	var riderId uuid.UUID
	var password_hash string

	err := r.db.QueryRowContext(ctx, query, req.Email).Scan(&password_hash, &riderId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginRepResponse{
				Rider_id:   uuid.Nil,
				Err:        status.Error(codes.Unauthenticated, "Invalid email or password"),
				Msg:        fmt.Sprintf("Login failed: email '%s' not found in database", req.Email),
				Status:     "Unauthenticated",
				StatusCode: uint32(codes.Unauthenticated),
			}
		}

		return LoginRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during login"),
			Msg:        fmt.Sprintf("SQL query failed: %v", err),
			Status:     "Internal server error",
			StatusCode: uint32(codes.Internal),
		}
	}

	isOk := utils.CheckPasswordHash(req.Password, password_hash)

	if !isOk {
		return LoginRepResponse{
			Rider_id:   uuid.Nil,
			Err:        status.Error(codes.Unauthenticated, "Invalid email or password"),
			Msg:        fmt.Sprintf("Login failed: password mismatch for email '%s'", req.Email),
			Status:     "Unauthenticated",
			StatusCode: uint32(codes.Unauthenticated),
		}
	}

	return LoginRepResponse{
		Rider_id:   riderId,
		Err:        nil,
		Msg:        "logged in successfully",
		Status:     "Ok",
		StatusCode: uint32(codes.OK),
	}

}

func (r *RiderAuthRepo) GetRiderInfoRepo(ctx context.Context, rider_id string) GetRiderInfoRepResponse {
	query := `SELECT firstname, lastname, email, phone_number, image_url, rating FROM rider_info WHERE rider_id = $1 LIMIT 1`

	var rider_info GetRiderInfo

	err := r.db.QueryRowContext(ctx, query, rider_id).Scan(
		&rider_info.Firstname,
		&rider_info.Lastname,
		&rider_info.Email,
		&rider_info.Phone_number,
		&rider_info.ImageURL,
		&rider_info.Rating,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetRiderInfoRepResponse{
				Err:        status.Error(codes.Unauthenticated, "Unauthorized"),
				Msg:        fmt.Sprintf("Getting Rider_info failed: rider_id '%s' not found in database", err),
				Status:     "Unauthenticated",
				StatusCode: uint32(codes.Unauthenticated),
			}
		}

		return GetRiderInfoRepResponse{
			Err:        status.Error(codes.Internal, "Internal server error"),
			Msg:        fmt.Sprintf("SQL query failed: %v", err),
			Status:     "Internal server error",
			StatusCode: uint32(codes.Internal),
		}
	}

	return GetRiderInfoRepResponse{
		Rider_info: GetRiderInfo{
			Rider_id:     rider_id,
			Firstname:    rider_info.Firstname,
			Lastname:     rider_info.Lastname,
			Phone_number: rider_info.Phone_number,
			Email:        rider_info.Email,
			ImageURL:     rider_info.ImageURL,
			Rating:       rider_info.Rating,
		},
		Err:        nil,
		Msg:        "Success",
		Status:     "Ok",
		StatusCode: 200,
	}
}

package driver_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth_service/internal/utils"
	"auth_service/proto/pb"
)

type DriverAuthRepo struct {
	db *sql.DB
}

func NewDriverAuthRepo(db *sql.DB) *DriverAuthRepo {
	return &DriverAuthRepo{
		db: db,
	}
}

func (d *DriverAuthRepo) RegisterDriver(ctx context.Context, req *pb.DriverRegisterRequest) RegisterRepResponse {
	driverInfo := req.DriverInfo
	carInfo := req.CarInfo

	if driverInfo == nil {
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Driver registration details are required"),
			Msg:        "Validation failed: DriverInfo is nil",
			Status:     "Bad Request",
			StatusCode: 400,
		}
	}

	if carInfo == nil {
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Car registration details are required"),
			Msg:        "Validation failed: CarInfo is nil",
			Status:     "Bad Request",
			StatusCode: 400,
		}
	}

	passwordHash, err := utils.HashPassword(driverInfo.Password)
	if err != nil {
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during registration"),
			Msg:        fmt.Sprintf("Crypto error: failed to hash password: %v", err),
			Status:     "Internal Server Error",
			StatusCode: 500,
		}
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during registration"),
			Msg:        fmt.Sprintf("Database transaction error: failed to begin transaction: %v", err),
			Status:     "Internal Server Error",
			StatusCode: 500,
		}
	}

	defer func() {
		_ = tx.Rollback()
	}()

	var driverID uuid.UUID
	driverQuery := `
		INSERT INTO driver_info (
			firstname, lastname, email, password_hash, phone_number, license_no, image_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING driver_id`

	err = tx.QueryRowContext(
		ctx,
		driverQuery,
		driverInfo.Firstname,
		driverInfo.Lastname,
		driverInfo.Email,
		passwordHash,
		driverInfo.PhoneNumber,
		driverInfo.LicenseNo,
		driverInfo.ImageUrl,
	).Scan(&driverID)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			clientMsg := "The provided information is already registered"

			if pqErr.Constraint == "driver_info_email_key" {
				clientMsg = "The email address is already registered"
			} else if pqErr.Constraint == "driver_info_license_no_key" {
				clientMsg = "The driving license number is already registered"
			} else if pqErr.Constraint == "driver_info_phone_number_key" {
				clientMsg = "The phone number is already registered"
			}

			return RegisterRepResponse{
				Driver_id:  uuid.Nil,
				Err:        status.Error(codes.AlreadyExists, clientMsg),
				Msg:        fmt.Sprintf("Conflict: unique constraint violation on %s", pqErr.Constraint),
				Status:     "Conflict",
				StatusCode: uint32(codes.AlreadyExists),
			}
		}
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during registration"),
			Msg:        fmt.Sprintf("SQL execution error: failed to insert driver_info: %v", err),
			Status:     "Internal Server Error",
			StatusCode: 500,
		}
	}

	vehicleQuery := `
		INSERT INTO vehicle_info (
			carname, brand, model, make, year, color, car_plate, insurance_policy_no, driver_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.ExecContext(
		ctx,
		vehicleQuery,
		carInfo.Carname,
		carInfo.Brand,
		carInfo.Model,
		carInfo.Make,
		carInfo.Year,
		carInfo.Color,
		carInfo.CarPlate,
		carInfo.InsurancePolicyNo,
		driverID,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			clientMsg := "The vehicle details are already registered"

			if pqErr.Constraint == "vehicle_info_car_plate_key" {
				clientMsg = "The vehicle license plate is already registered"
			} else if pqErr.Constraint == "vehicle_info_insurance_policy_no_key" {
				clientMsg = "The insurance policy number is already registered"
			}

			return RegisterRepResponse{
				Driver_id:  uuid.Nil,
				Err:        status.Error(codes.AlreadyExists, clientMsg),
				Msg:        fmt.Sprintf("Conflict: unique constraint violation on %s", pqErr.Constraint),
				Status:     "Conflict",
				StatusCode: uint32(codes.AlreadyExists),
			}
		}
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during registration"),
			Msg:        fmt.Sprintf("SQL execution error: failed to insert vehicle_info: %v", err),
			Status:     "Internal Server Error",
			StatusCode: 500,
		}
	}

	err = tx.Commit()
	if err != nil {
		return RegisterRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred finalizing your registration"),
			Msg:        fmt.Sprintf("Database transaction error: failed to commit transaction: %v", err),
			Status:     "Internal Server Error",
			StatusCode: 500,
		}
	}

	return RegisterRepResponse{
		Driver_id:  driverID,
		Err:        nil,
		Msg:        "Success: Driver account and vehicle registered successfully",
		Status:     "Ok",
		StatusCode: 201,
	}
}

func (d *DriverAuthRepo) LoginDriver(ctx context.Context, req *pb.DriverLoginRequest) LoginRepResponse {

	if req.Email == "" {
		return LoginRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Driver email is required"),
			Msg:        "Validation failed: email is nil",
			Status:     "Bad Request",
			StatusCode: 400,
		}
	}

	if req.Password == "" {
		return LoginRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.InvalidArgument, "Driver password is required"),
			Msg:        "Validation failed: password is nil",
			Status:     "Bad Request",
			StatusCode: 400,
		}
	}

	query := `SELECT password_hash, driver_id FROM driver_info WHERE email =$1`

	var driver_id uuid.UUID
	var password_hash string

	err := d.db.QueryRowContext(ctx, query, req.Email).Scan(&password_hash, &driver_id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginRepResponse{
				Driver_id:  uuid.Nil,
				Err:        status.Error(codes.Unauthenticated, "Invalid email or password"),
				Msg:        fmt.Sprintf("Login failed: email '%s' not found in database", req.Email),
				Status:     "Unauthenticated",
				StatusCode: uint32(codes.Unauthenticated),
			}
		}

		return LoginRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Internal, "An unexpected error occurred during login"),
			Msg:        fmt.Sprintf("SQL query failed: %v", err),
			Status:     "Internal server error",
			StatusCode: uint32(codes.Internal),
		}
	}

	isOk := utils.CheckPasswordHash(req.Password, password_hash)

	if !isOk {
		return LoginRepResponse{
			Driver_id:  uuid.Nil,
			Err:        status.Error(codes.Unauthenticated, "Invalid email or password"),
			Msg:        fmt.Sprintf("Login failed: password mismatch for email '%s'", req.Email),
			Status:     "Unauthenticated",
			StatusCode: uint32(codes.Unauthenticated),
		}
	}

	return LoginRepResponse{
		Driver_id:  driver_id,
		Err:        nil,
		Msg:        "logged in successfully",
		Status:     "Ok",
		StatusCode: uint32(codes.OK),
	}

}

func (d *DriverAuthRepo) GetDriverInfoRepo(ctx context.Context, driver_id string) GetDriverInfoRepResponse {
	query := `SELECT firstname, lastname, email, phone_number, image_url, license_no, rating FROM driver_info WHERE driver_id = $1 LIMIT 1`

	var driverinfo GetDriveInfo

	err := d.db.QueryRowContext(ctx, query, driver_id).Scan(&driverinfo.Firstname, &driverinfo.Lastname, &driverinfo.Email, &driverinfo.Phone_number, &driverinfo.ImageURL, &driverinfo.License_number, &driverinfo.Rating)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetDriverInfoRepResponse{
				Err:        status.Error(codes.Unauthenticated, "Unauthorized"),
				Msg:        fmt.Sprintf("Getting Driver_info failed: driver_id '%s' not found in database", err),
				Status:     "Unauthenticated",
				StatusCode: uint32(codes.Unauthenticated),
			}
		}

		return GetDriverInfoRepResponse{
			Err:        status.Error(codes.Internal, "Internal server error"),
			Msg:        fmt.Sprintf("SQL query failed: %v", err),
			Status:     "Internal server error",
			StatusCode: uint32(codes.Internal),
		}
	}

	return GetDriverInfoRepResponse{
		Driver_info: GetDriveInfo{
			Driver_id:      driverinfo.Driver_id,
			Firstname:      driverinfo.Firstname,
			Lastname:       driverinfo.Lastname,
			Phone_number:   driverinfo.Phone_number,
			Email:          driverinfo.Email,
			ImageURL:       driverinfo.ImageURL,
			License_number: driverinfo.License_number,
			Rating:         driverinfo.Rating,
		},
		Err:        nil,
		Msg:        "Success",
		Status:     "Ok",
		StatusCode: 200,
	}
}

func (d *DriverAuthRepo) GetDriverCarInfoRepo(driver_id string, ctx context.Context) (Vehicle, error) {
	query := `SELECT
    vehicle_id,
    carname,
    brand,
    model,
    make,
    year,
    color,
    car_plate,
    insurance_policy_no,
    is_verified
	FROM vehicle_info
	WHERE driver_id = $1;`

	var carinfo Vehicle

	err := d.db.QueryRowContext(ctx, query, driver_id).Scan(&carinfo.VehicleID, &carinfo.CarName, &carinfo.Brand, &carinfo.Model, &carinfo.Make, &carinfo.Year, &carinfo.Color, &carinfo.CarPlate, &carinfo.InsurancePolicyNo, &carinfo.IsVerified)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Vehicle{}, status.Error(codes.Unauthenticated, "Unauthorized")
		}
		return Vehicle{}, status.Error(codes.Internal, "Internal server error")
	}

	return carinfo, nil
}

func (d *DriverAuthRepo) UpdatePassword(ctx context.Context, driverid string, newPassword string, cur_password string) error {
	query := `SELECT password_hash FROM driver_info WHERE driver_id = $1`

	var passoword_hash string

	err := d.db.QueryRowContext(ctx, query, driverid).Scan(&passoword_hash)

	if err != nil {
		fmt.Println("db query looking for password failed", err)
		return status.Error(codes.Internal, "Internal server error")
	}

	isMatch := utils.CheckPasswordHash(cur_password, passoword_hash)

	if !isMatch {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}

	new_password_hash, err := utils.HashPassword(newPassword)

	if err != nil {
		fmt.Println("Error :: Hashing password", err)
		return status.Error(codes.Internal, "Internal server error")
	}

	query = `
	UPDATE driver_info
	SET password_hash = $1
		updated_at = NOW()
	WHERE driver_id = $2
 	`

	_, err = d.db.Exec(query, new_password_hash, driverid)
	if err != nil {
		fmt.Println("Error :: updating driver password", err)
		return status.Error(codes.Internal, "Internal server error")
	}

	return nil
}

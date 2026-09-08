package driver_service

import "github.com/google/uuid"

type RegisterResponse struct {
	Message       string
	Err           error
	Access_token  string
	Refresh_token string
	Status        string
	StatusCode    int16
}

type LoginResponse struct {
	Message       string
	Err           error
	Access_token  string
	Refresh_token string
	Status        string
	StatusCode    int16
}

type GetDriveInfo struct {
	Driver_id      string
	Firstname      string
	Lastname       string
	Phone_number   string
	Email          string
	ImageURL       string
	License_number string
	Rating         float64
}

type GetDriverInfoServiceResponse struct {
	Driver_info GetDriveInfo
	Message     string
	Err         error
	Status      string
	StatusCode  int16
}


type Vehicle struct {
	VehicleID         uuid.UUID `db:"vehicle_id"`
	CarName           string    `db:"carname"`
	Brand             string    `db:"brand"`
	Model             string    `db:"model"`
	Make              string    `db:"make"`
	Year              int       `db:"year"`
	Color             string    `db:"color"`
	CarPlate          string    `db:"car_plate"`
	InsurancePolicyNo string    `db:"insurance_policy_no"`
	IsVerified        bool      `db:"is_verified"`
}

package driver_repository

import (
	"github.com/google/uuid"
)

type RegisterRepResponse struct {
	Driver_id  uuid.UUID
	Err        error
	Msg        string
	Status     string
	StatusCode uint32
}

type LoginRepResponse struct {
	Driver_id  uuid.UUID
	Err        error
	Msg        string
	Status     string
	StatusCode uint32
}

type GetDriveInfo struct {
	Driver_id      uuid.UUID
	Firstname      string
	Lastname       string
	Phone_number   string
	Email          string
	ImageURL       string
	License_number string
	Rating         float64
}

type GetDriverInfoRepResponse struct {
	Driver_info GetDriveInfo
	Err         error
	Msg         string
	Status      string
	StatusCode  uint32
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

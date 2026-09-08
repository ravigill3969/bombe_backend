package rider_repository

import "github.com/google/uuid"

type RegisterRepResponse struct {
	Rider_id  uuid.UUID
	Err        error
	Msg        string
	Status     string
	StatusCode uint32
}

type LoginRepResponse struct {
	Rider_id  uuid.UUID
	Err        error
	Msg        string
	Status     string
	StatusCode uint32
}

type GetRiderInfo struct {
	Rider_id      string
	Firstname      string
	Lastname       string
	Phone_number   string
	Email          string
	ImageURL       string
	Rating         float64
}

type GetRiderInfoRepResponse struct {
	Rider_info GetRiderInfo
	Err         error
	Msg         string
	Status      string
	StatusCode  uint32
}

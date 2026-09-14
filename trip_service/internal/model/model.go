package model

import "time"

type SQSDataForCreatingTrip struct {
	Status      int32  `json:"status"`
	Aud         string `json:"aud"`
	Message     string `json:"message"`
	TempRideID  string `json:"temp_ride_id"`
	PaymentID   string `json:"payment_id"`
	RiderID     string `json:"rider_id"`
	ServiceType string `json:"service_type"`

	Pickup      Location    `json:"pickup"`
	Dropoff     Location    `json:"dropoff"`
	Fare        Fare        `json:"fare"`
	RideDetails RideDetails `json:"ride_details"`
}

type Location struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Fare struct {
	AmountInCents int64  `json:"amount_in_cents"`
	Currency      string `json:"currency"`
}

type RideDetails struct {
	DistanceMeters  int64  `json:"distance_meters"`
	DurationSeconds int64  `json:"duration_seconds"`
	ServiceType     string `json:"service_type"`
}

type SQSDataToMainServer struct {
	Status      int32  `json:"status"`
	ServerType  string `json:"server_type"`
	Aud         string `json:"aud"`
	Message     string `json:"message"`
	TripID      string `json:"trip_id"`
	PaymentID   string `json:"payment_id"`
	RiderID     string `json:"rider_id"`
	DriverFare  int64  `json:"driver_fare"`
	For         string `json:"for"`
	RideDetails RideDetails
	Pickup      Location
	Dropoff     Location
}

type AcceptDriverTripParams struct {
	Rider_id  string
	Driver_id string
	Trip_id   string
}

type Trip struct {
	TripID             string
	RiderID            string
	DriverID           *string
	PaymentID          string
	Status             string
	PickupAddress      string
	PickupLatitude     float64
	PickupLongitude    float64
	DropoffAddress     string
	DropoffLatitude    float64
	DropoffLongitude   float64
	Fare               float64
	DriverAmount       float32
	PlatformFee        float64
	TaxAmount          float64
	DiscountAmount     float64
	DistanceMeters     int64
	DurationSeconds    int64
	CancellationReason *string
	CancelledBy        *string
	CreatedAt          time.Time
	AcceptedAt         *time.Time
	PickedUpAt         *time.Time
	CompletedAt        *time.Time
	CancelledAt        *time.Time
	UpdatedAt          time.Time
}

type SQSTripCancelRequestToPayment struct {
	PaymentId  string `json:"payment_id"`
	ServerType string `json:"server_type"`
	Aud        string `json:"aud"`
	Message    string `json:"message"`
	RiderID    string `json:"rider_id"`
	DriverId   string `json:"driver_id"`
}

type SQSTripCancelRequestToMain struct {
	PaymentId           string `json:"payment_id"`
	ServerType          string `json:"server_type"`
	Aud                 string `json:"aud"`
	Message             string `json:"message"`
	RiderID             string `json:"rider_id"`
	DriverId            string `json:"driver_id"`
	For                 string `json:"for"`
	IsCancelledByDriver bool   `json:"is_cancelled_by_driver"`
	TripId              string `json:"trip_id"`
}

type SQSTripCompletedRequestToMain struct {
	ServerType          string `json:"server_type"`
	Aud                 string `json:"aud"`
	Message             string `json:"message"`
	RiderID             string `json:"rider_id"`
	DriverId            string `json:"driver_id"`
	For                 string `json:"for"`
	TripId              string `json:"trip_id"`
}
package model

type GeoLocation struct {
	Latitude  float64
	Longitude float64
	Address   string
}

type FareDetails struct {
	AmountInCents int64
	Currency      string
}

type RideDetails struct {
	DurationSeconds int64
	DistanceMeters  int64
}

type CreateCheckoutParams struct {
	ServiceType string
	Pickup      GeoLocation
	Dropoff     GeoLocation
	Fare        FareDetails
	Ride        RideDetails
}
type CheckoutSessionResult struct {
	CheckoutURL     string
	StripeSessionID string
	TemporaryRideID string
}

type PaymentSuccessParams struct {
	ServiceType string	
	Pickup      GeoLocation
	Dropoff     GeoLocation
	TempRideId  string
	Fare        FareDetails
	Ride        RideDetails
	PaymentIntentId string
}

type SQSTripCancelRequestFromCancelTrip struct {
	PaymentId  string `json:"payment_id"`
	ServerType string `json:"server_type"`
	Aud        string `json:"aud"`
	Message    string `json:"message"`
	RiderID    string `json:"rider_id"`
	DriverId   string `json:"driver_id"`
}
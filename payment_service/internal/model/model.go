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
}

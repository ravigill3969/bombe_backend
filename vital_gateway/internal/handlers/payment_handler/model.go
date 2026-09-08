package payment_handler

type GeoLocation struct {
	Latitude  float64
	Longitude float64
	Address   string
}

type ServiceType int32

const (
	ServiceTypeUnspecified ServiceType = iota
	ServiceTypeEconomy
	ServiceTypeComfort
	ServiceTypeXL
)

type FareDetails struct {
	AmountInCents int64
	Currency      string
}

type CreateCheckoutSessionRequest struct {
	ServiceType ServiceType
	Pickup      GeoLocation
	Dropoff     GeoLocation
	Fare        FareDetails
}

type CreateCheckoutSessionResponse struct {
	CheckoutURL     string
	StripeSessionID string
	TemporaryRideID string
}

type RideMetadata struct {
	RiderID         string
	ServiceType     string
	PickupAddress   string
	DropoffAddress  string
	PickupLat       float64
	PickupLng       float64
	DropoffLat      float64
	DropoffLng      float64
	Amount          int64
	TempRideId      string
	DurationSeconds int64
	DistanceMeters  int64
}

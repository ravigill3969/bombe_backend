package trip

type SQSDataToMainServer struct {
	Status      int32  `json:"status"`
	ServerType  string `json:"server_type"`
	Aud         string `json:"aud"`
	Message     string `json:"message"`
	TripID      string `json:"trip_id"`
	PaymentID   string `json:"payment_id"`
	RiderID     string `json:"rider_id"`
	DriverFare  int64  `json:"driver_fare"`
	RideDetails RideDetails
	Pickup      Location
	Dropoff     Location
}
type RideDetails struct {
	DistanceMeters  int64  `json:"distance_meters"`
	DurationSeconds int64  `json:"duration_seconds"`
	ServiceType     string `json:"service_type"`
}
type Location struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TripReqToDriver struct {
	Type        string      `json:"type"`
	DriverFare  int64       `json:"driver_fare"`
	RideDetails RideDetails `json:"ride_details"`
	Pickup      Location    `json:"pickup"`
	Dropoff     Location    `json:"dropoff"`
	RiderID     string      `json:"rider_id"`
	TripID      string      `json:"trip_id"`
}

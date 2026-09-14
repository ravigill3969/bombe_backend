package trip

type SQSDataFor struct {
	For string `json:"for"` // this is to check to check what this msg is meant for?
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

type SQSTripCancelRequestToMain struct {
	PaymentId           string `json:"payment_id"`
	ServerType          string `json:"server_type"`
	Aud                 string `json:"aud"`
	Message             string `json:"message"`
	RiderId             string `json:"rider_id"`
	DriverId            string `json:"driver_id"`
	For                 string `json:"for"`
	IsCancelledByDriver bool   `json:"is_cancelled_by_driver"`
	TripId              string `json:"trip_id"`
}

type CancelTripDataToUser struct {
	Type          string `json:"type"` //type of message so ws can handle properly in frentend
	TripId        string `json:"trip_id"`
	RiderID       string `json:"rider_id"`
	DriverId      string `json:"driver_id"`
	Message       string `json:"message"`
	DriverOrRider string `json:"driver_or_rider"`
}

type SQSTripCompletedRequestToMain struct {
	ServerType string `json:"server_type"`
	Aud        string `json:"aud"`
	Message    string `json:"message"`
	RiderId    string `json:"rider_id"`
	DriverId   string `json:"driver_id"`
	For        string `json:"for"`
	TripId     string `json:"trip_id"`
}

type CompleteTripDataToUser struct {
	Type     string `json:"type"` //type of message so ws can handle properly in frentend
	TripId   string `json:"trip_id"`
	RiderId  string `json:"rider_id"`
	DriverId string `json:"driver_id"`
}

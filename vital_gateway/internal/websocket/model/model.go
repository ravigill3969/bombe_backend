package socket_model

type DataToSendOnLive struct {
	Type      string  `json:"type"`
	DriverID  string  `json:"driver_id"`
	RiderID   string  `json:"rider_id,omitempty"`
	RideID    string  `json:"ride_id,omitempty"`
	IsOnline  bool    `json:"is_online,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Status    string  `json:"status,omitempty"`
}

type DataToSendTORiderFromDriver struct {
	Type      string  `json:"type"`
	DriverID  string  `json:"driver_id"`
	RiderID   string  `json:"rider_id,omitempty"`
	RideID    string  `json:"ride_id,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	TripID    string  `json:"trip_id,omitempty"`
	Status    string  `json:"status"`
}

package redis_handler

type GeoLocation struct {
	Latitude  float64
	Longitude float64
}

type DriverInfo struct {
	DriverID string
	IsOnline bool
	RideID   string
}
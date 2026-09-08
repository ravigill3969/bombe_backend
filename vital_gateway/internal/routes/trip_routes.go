package routes

import (
	"bombe_main_server/internal/handlers/trip_handler"
	"bombe_main_server/internal/middleware/driver_middleware"
	"bombe_main_server/internal/middleware/rider_middleware"
	"net/http"
)

type TripRoutes struct {
	mux     *http.ServeMux
	handler *trip_handler.TripHandler
}

func NewTripRoutes(mux *http.ServeMux, handler *trip_handler.TripHandler) *TripRoutes {
	return &TripRoutes{
		mux:     mux,
		handler: handler,
	}
}

func (r *TripRoutes) Register() {
	r.mux.HandleFunc("GET /api/trip/get-active-trip-with-driverid", driver_middleware.DriverAuthMiddleware(r.handler.GetTripWithDriverId))
	r.mux.HandleFunc("GET /api/trip/get-active-trip-with-riderid", rider_middleware.RiderAuthMiddleware(r.handler.GetTripWithRiderId))

	r.mux.HandleFunc("POST /api/trip/assign-driver", driver_middleware.DriverAuthMiddleware(r.handler.AssignTripToDriver))
	r.mux.HandleFunc("POST /api/trip/rider-picked", driver_middleware.DriverAuthMiddleware(r.handler.UpdateTripStatusToPicked))
	r.mux.HandleFunc("POST /api/trip/trip-completed", driver_middleware.DriverAuthMiddleware(r.handler.MarkTripCompleted))

}

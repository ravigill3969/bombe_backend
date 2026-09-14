package routes

import (
	"bombe_main_server/internal/handlers/rider_handler"
	"bombe_main_server/internal/middleware/rider_middleware"
	"net/http"
)

type RiderAuthRoutes struct {
	mux     *http.ServeMux
	handler *rider_handler.AuthHandler
}

func NewRiderAuthRoutes(mux *http.ServeMux, handler *rider_handler.AuthHandler) *RiderAuthRoutes {
	return &RiderAuthRoutes{
		mux:     mux,
		handler: handler,
	}
}

func (r *RiderAuthRoutes) Register() {
	r.mux.HandleFunc("POST /api/rider/register", r.handler.RegisterRider)
	r.mux.HandleFunc("POST /api/rider/login", r.handler.LoginRider)
	r.mux.HandleFunc("POST /api/rider/update-password", rider_middleware.RiderAuthMiddleware(r.handler.UpdatePassword))

	r.mux.HandleFunc("GET /api/rider/verify", rider_middleware.RiderAuthMiddleware(r.handler.GetRiderInfoo))

	r.mux.HandleFunc("GET /api/rider/logout", rider_middleware.RiderAuthMiddleware(r.handler.LogoutRider))
}

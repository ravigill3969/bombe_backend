package routes

import (
	"bombe_main_server/internal/handlers/driver_handler"
	"bombe_main_server/internal/middleware/driver_middleware"
	"net/http"
)

type DriverAuthRoutes struct {
	mux     *http.ServeMux
	handler *driver_handler.AuthHandler
}

func NewDriverAuthRoutes(mux *http.ServeMux, handler *driver_handler.AuthHandler) *DriverAuthRoutes {
	return &DriverAuthRoutes{
		mux:     mux,
		handler: handler,
	}
}

func (d *DriverAuthRoutes) Register() {
	d.mux.HandleFunc("POST /api/driver/login", d.handler.LoginDriver)
	d.mux.HandleFunc("POST /api/driver/register", d.handler.RegisterDriver)
	d.mux.HandleFunc("POST /api/driver/update-password", driver_middleware.DriverAuthMiddleware(d.handler.UpdatePassword))
	d.mux.HandleFunc("POST /api/driver/logout", driver_middleware.DriverAuthMiddleware(d.handler.LogoutDriver))

	d.mux.HandleFunc("GET /api/driver/verify", driver_middleware.DriverAuthMiddleware(d.handler.GetDriverInfoo))
	d.mux.HandleFunc("GET /api/driver/carinfo", driver_middleware.DriverAuthMiddleware(d.handler.GetDriverCarInfoo))

}

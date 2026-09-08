package routes

import (
	"bombe_main_server/internal/handlers/payment_handler"
	"bombe_main_server/internal/middleware/rider_middleware"
	"net/http"
)

type PaymentRoutes struct {
	mux     *http.ServeMux
	handler *payment_handler.PaymentHandler
}

func NewPaymentRoutes(mux *http.ServeMux, handler *payment_handler.PaymentHandler) *PaymentRoutes {
	return &PaymentRoutes{
		mux:     mux,
		handler: handler,
	}
}

func (p *PaymentRoutes) Register() {
	p.mux.HandleFunc("POST /api/payment/create-checkout-session", rider_middleware.RiderAuthMiddleware(p.handler.CreateCheckOutSession))
	p.mux.HandleFunc("POST /api/payment/webhook", p.handler.StripeWebhook)
}

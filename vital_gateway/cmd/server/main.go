package main

import (
	"bombe_main_server/internal/awssqs"
	"bombe_main_server/internal/client/driver_client"
	"bombe_main_server/internal/client/payment_client"
	"bombe_main_server/internal/client/rider_client"
	tripclient "bombe_main_server/internal/client/trip_client"
	"bombe_main_server/internal/handlers/driver_handler"
	"bombe_main_server/internal/handlers/payment_handler"
	"bombe_main_server/internal/handlers/rider_handler"
	"bombe_main_server/internal/handlers/trip_handler"
	redis_handler "bombe_main_server/internal/redis"
	"bombe_main_server/internal/routes"
	"bombe_main_server/internal/trip"

	sqscfg "github.com/aws/aws-sdk-go-v2/config"

	websocket_handler "bombe_main_server/internal/websocket"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading env: %v", err)
	// }

	PORT := os.Getenv("PORT")

	if PORT == "" {
		log.Fatalln("PORT env is required")
	}

	UPSTASH_REDIS_PROPER_URL := os.Getenv("UPSTASH_REDIS_PROPER_URL")

	if UPSTASH_REDIS_PROPER_URL == "" {
		log.Fatalln("UPSTASH_REDIS_PROPER_URL env is required")
	}

	AWS_SQS_URL := os.Getenv("AWS_SQS_URL")

	if AWS_SQS_URL == "" {
		log.Fatalln("AWS_SQS_URL env is required")
	}

	AWS_REGION := os.Getenv("AWS_REGION")

	if AWS_REGION == "" {
		log.Fatalln("AWS_REGION env is required")
	}

	sqs_client := InitSQS(AWS_REGION)
	sqs_queue := awssqs.NewSQSQueue(sqs_client, AWS_SQS_URL)

	redis := redis_handler.NewRedisHandler(UPSTASH_REDIS_PROPER_URL)

	drivergrpc_client, drivergrpc_conn := driver_client.DriverClient()
	defer drivergrpc_conn.Close()
	driver_handler := driver_handler.NewAuthHandler(drivergrpc_client)

	ridergrpc_client, ridergrpc_conn := rider_client.RiderClinet()
	defer ridergrpc_conn.Close()
	rider_handler := rider_handler.NewAuthHandler(ridergrpc_client)

	paymentgrpc_client, paymentgrpc_conn := payment_client.PaymentClient()
	defer paymentgrpc_conn.Close()
	payment_handler := payment_handler.NewPaymentHandler(paymentgrpc_client)

	tripgrpc_client, tripgrpc_conn := tripclient.TripClient()
	defer tripgrpc_conn.Close()
	trip_handler := trip_handler.NewTripHandler(tripgrpc_client)

	mux := http.NewServeMux()

	driver_routes := routes.NewDriverAuthRoutes(mux, driver_handler)
	driver_routes.Register()

	rider_routes := routes.NewRiderAuthRoutes(mux, rider_handler)
	rider_routes.Register()

	payment_routes := routes.NewPaymentRoutes(mux, payment_handler)
	payment_routes.Register()

	trip_routes := routes.NewTripRoutes(mux, trip_handler)
	trip_routes.Register()

	wsHandler := websocket_handler.NewWSHandler()
	trip := trip.NewTripHandler(redis, sqs_queue, wsHandler)

	mux.HandleFunc("/ws", wsHandler.WSHandler)
	fmt.Println("Starting server on PORT=", PORT)

	//starting channel here
	go trip.ReadMessagesStoredInChannel()
	go trip.StartListeningToSQS()

	err := http.ListenAndServe(PORT, EnableCorsWithCredentials(mux))

	if err != nil {
		log.Fatalf("error starting server %s", err)
	}
}

func EnableCorsWithCredentials(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"https://booombe.com":       true,
			"https://www.booombe.com":   true,
			"https://iloverher.com":     true,
			"https://www.iloverher.com": true,
			"http://localhost:5173":     true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		w.Header().Set("Access-Control-Allow-Private-Network", "true")
		w.Header().Set("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func InitSQS(AWS_REGION string) *sqs.Client {
	cfg, err := sqscfg.LoadDefaultConfig(context.TODO(), sqscfg.WithRegion(AWS_REGION))
	if err != nil {
		panic(err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	return sqsClient
}

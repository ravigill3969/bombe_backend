package tripclient

import (
	"fmt"
	"log"
	"os"
	"trip_service/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TripClient() (pb.TripServiceClient, *grpc.ClientConn) {
	trip_port := os.Getenv("TRIP_PORT")

	if trip_port == "" {
		trip_port = "trip-service:50052"
	}

	client_url := fmt.Sprintf(trip_port)

	conn, err := grpc.NewClient(client_url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to gRPC backend: %v", err)
	}

	client := pb.NewTripServiceClient(conn)

	return client, conn
}

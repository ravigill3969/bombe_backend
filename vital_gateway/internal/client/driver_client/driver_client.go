package driver_client

import (
	"auth_service/proto/pb"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DriverClient() (pb.DriverAuthServiceClient, *grpc.ClientConn) {
	auth_port := os.Getenv("AUTH_PORT")

	if auth_port == "" {
		auth_port = "auth-service:50052"
	}

	client_url := fmt.Sprintf(auth_port)

	conn, err := grpc.NewClient(client_url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to gRPC backend: %v", err)
	}

	client := pb.NewDriverAuthServiceClient(conn)

	return client, conn
}

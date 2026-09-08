package payment_client

import (
	"fmt"
	"log"
	"os"
	"payment_service/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func PaymentClient() (pb.PaymentServiceClient, *grpc.ClientConn) {
	payment_port := os.Getenv("PAYMENT_PORT")

	if payment_port == "" {
		payment_port = "payment-service:50052"
	}

	client_url := fmt.Sprintf(payment_port)

	conn, err := grpc.NewClient(client_url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to gRPC backend: %v", err)
	}

	client := pb.NewPaymentServiceClient(conn)

	return client, conn
}

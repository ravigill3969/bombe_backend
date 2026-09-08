package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"auth_service/internal/config"
	pb "auth_service/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	isOk := config.LoadEnv()

	if !isOk {
		return
	}

	PORT := os.Getenv("PORT")
	config.IsEnvSet("PORT", PORT)

	client_addr := fmt.Sprintf("%s%s", "localhost", PORT)

	conn, err := grpc.NewClient(client_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC server at localhost%s: %v", PORT, err)
	}
	defer conn.Close()

	cd := pb.NewDriverAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := cd.LoginDriver(ctx, &pb.DriverLoginRequest{
		// Pass fields here if your .proto file requires them, for example:
		// Username: "ravi",
		Email: "water",
	})
	if err != nil {
		log.Fatalf("error calling function Login: %v", err)
	}

	fmt.Println(r.StatusCode, r.Message)

	cr := pb.NewRiderAuthServiceClient(conn)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := cr.LoginRider(ctx, &pb.RiderLoginRequest{
		Email:    "wa",
		Password: "dfewv",
	})

	fmt.Println(res.Message, res.StatusCode)

}

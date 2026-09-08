package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"trip_service/internal/config"
	"trip_service/internal/middleware"
	"trip_service/internal/server"
	"trip_service/internal/service"
	repository "trip_service/internal/tepository"
	"trip_service/proto/pb"

	sqscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"google.golang.org/grpc"
)

func main() {
	isOk := config.LoadEnv()

	if !isOk {
		return
	}

	PORT := os.Getenv("PORT")

	if PORT == "" {
		log.Fatalln("PORT env is required")
	}

	lis, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatalf("failed to listen on port %s %v", PORT, err)
	}

	db := config.Connect_to_db()
	defer db.Close()

	AWS_SQS_URL := os.Getenv("AWS_SQS_URL")

	if AWS_SQS_URL == "" {
		log.Fatalln("AWS_SQS_URL env is required")
	}

	sqsclient := InitSQS()
	sqsRepo := repository.NewSQSQueue(sqsclient, AWS_SQS_URL)

	s := grpc.NewServer(grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor))

	pb.RegisterTripServiceServer(s, InitTripServer(db, sqsRepo))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func InitTripServer(
	db *sql.DB,
	sqsrepo *repository.SQSQueue,
) *server.TripServer {
	tripRepo := repository.NewTripRepo(db)
	tripService := service.NewTripService(sqsrepo, tripRepo)
	tripServer := server.NewTripServer(tripService)

	go tripService.StartListeningSQS(context.Background())

	return tripServer
}

func InitSQS() *sqs.Client {
	cfg, err := sqscfg.LoadDefaultConfig(
		context.TODO(),
		sqscfg.WithRegion("us-east-1"))
	if err != nil {
		panic(err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	return sqsClient
}

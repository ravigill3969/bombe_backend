package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"payment_service/internal/config"
	"payment_service/internal/middleware"
	"payment_service/internal/server"
	"payment_service/internal/service"
	repository "payment_service/internal/tepository"
	"payment_service/proto/pb"

	"google.golang.org/grpc"
)

func main() {
	isOk := config.LoadEnv()

	if !isOk {
		return
	}

	PORT := os.Getenv("PORT")

	if PORT == "" {
		fmt.Println("PORT env is required")
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

	sqsclient := config.InitSQS()
	sqsRepo := repository.NewSQSQueue(sqsclient, AWS_SQS_URL)

	s := grpc.NewServer(grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor))

	pb.RegisterPaymentServiceServer(s, InitPaymentServer(db, *sqsRepo))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func InitPaymentServer(db *sql.DB, sqsrepo repository.SQSQueue) *server.PaymentServer {
	paymentRepo := repository.NewPaymentRepo(db)
	paymentService := service.NewPaymentService(*paymentRepo, &sqsrepo)
	paymentServer := server.NewPaymentServer(*paymentService)
	return paymentServer
}


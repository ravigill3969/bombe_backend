package main

import (
	"auth_service/internal/config"
	driver_server "auth_service/internal/driver/server"
	driver_service "auth_service/internal/driver/service"
	driver_repository "auth_service/internal/driver/tepository"
	"auth_service/internal/middleware"
	rider_server "auth_service/internal/rider/server"
	rider_service "auth_service/internal/rider/service"
	rider_repository "auth_service/internal/rider/tepository"
	pb "auth_service/proto/pb"
	"database/sql"
	"log"
	"net"
	"os"

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

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GrpcAuthInterceptor),
	)

	pb.RegisterRiderAuthServiceServer(s, InitRiderServer(db))
	pb.RegisterDriverAuthServiceServer(s, InitDriverServer(db))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func InitRiderServer(db *sql.DB) *rider_server.RiderAuthServer {

	authRiderRepo := rider_repository.NewRiderAuthRepo(db)
	authRiderService := rider_service.NewRiderAuthService(authRiderRepo)
	authRiderServer := rider_server.NewRiderAuthServer(authRiderService)

	return authRiderServer
}

func InitDriverServer(db *sql.DB) *driver_server.DriverAuthServer {

	authDriverRepo := driver_repository.NewDriverAuthRepo(db)
	authDriverService := driver_service.NewDriverAuthService(authDriverRepo)
	authDriverServer := driver_server.NewDriverAuthServer(authDriverService)

	return authDriverServer
}

package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	pb "github.com/omnik/tech-ip-sem2/proto/auth"
	grpcserver "github.com/omnik/tech-ip-sem2/services/auth/internal/grpc"
	httphandler "github.com/omnik/tech-ip-sem2/services/auth/internal/http"
	"github.com/omnik/tech-ip-sem2/services/auth/internal/service"
	"github.com/omnik/tech-ip-sem2/shared/middleware"
)

func main() {
	httpPort := os.Getenv("AUTH_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}
	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	svc := service.New()

	// Запуск gRPC сервера
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("gRPC listen error: %v", err)
		}
		grpcSrv := grpc.NewServer()
		pb.RegisterAuthServiceServer(grpcSrv, grpcserver.New(svc))
		log.Printf("Auth gRPC server starting on :%s", grpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	// Запуск HTTP сервера
	handler := httphandler.New(svc)
	var mux http.Handler = handler.Routes()
	mux = middleware.Logging(mux)
	mux = middleware.RequestID(mux)

	log.Printf("Auth HTTP server starting on :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("Auth HTTP server failed: %v", err)
	}
}

package grpcserver

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/omnik/tech-ip-sem2/proto/auth"
	"github.com/omnik/tech-ip-sem2/services/auth/internal/service"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func New(svc *service.AuthService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	log.Printf("[gRPC] Verify request for token: %.10s...", req.Token)

	subject, valid := s.svc.Verify(req.Token)
	if !valid {
		log.Printf("[gRPC] Token verification failed: invalid token")
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	log.Printf("[gRPC] Token verified for subject: %s", subject)
	return &pb.VerifyResponse{
		Valid:   true,
		Subject: subject,
	}, nil
}

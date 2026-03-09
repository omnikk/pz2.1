package authclient

import (
"context"
"fmt"
"log"
"time"

"google.golang.org/grpc"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/credentials/insecure"
"google.golang.org/grpc/metadata"
"google.golang.org/grpc/status"

pb "github.com/omnik/tech-ip-sem2/proto/auth"
)

type GrpcClient struct {
client pb.AuthServiceClient
}

func NewGrpc(addr string) (*GrpcClient, error) {
conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
return nil, err
}
return &GrpcClient{client: pb.NewAuthServiceClient(conn)}, nil
}

func (c *GrpcClient) Verify(ctx context.Context, token, requestID string) (string, error) {
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()

if requestID != "" {
ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
}

log.Printf("[%s] Calling Auth gRPC verify", requestID)

resp, err := c.client.Verify(ctx, &pb.VerifyRequest{Token: token})
if err != nil {
st, _ := status.FromError(err)
if st.Code() == codes.Unauthenticated {
log.Printf("[%s] Auth gRPC verify: unauthorized", requestID)
return "", ErrUnauthorized
}
log.Printf("[%s] Auth gRPC verify failed: %v", requestID, err)
return "", fmt.Errorf("auth unavailable: %w", err)
}

log.Printf("[%s] Auth gRPC verify: success, subject=%s", requestID, resp.Subject)
return resp.Subject, nil
}

package main

import (
"log"
"net/http"
"os"

httphandler "github.com/omnik/tech-ip-sem2/services/tasks/internal/http"
"github.com/omnik/tech-ip-sem2/services/tasks/internal/client/authclient"
"github.com/omnik/tech-ip-sem2/services/tasks/internal/service"
"github.com/omnik/tech-ip-sem2/shared/middleware"
)

func main() {
port := os.Getenv("TASKS_PORT")
if port == "" {
port = "8082"
}

authMode := os.Getenv("AUTH_MODE")
if authMode == "" {
authMode = "grpc"
}

var auth httphandler.AuthVerifier

if authMode == "grpc" {
grpcAddr := os.Getenv("AUTH_GRPC_ADDR")
if grpcAddr == "" {
grpcAddr = "localhost:50051"
}
c, err := authclient.NewGrpc(grpcAddr)
if err != nil {
log.Fatalf("Failed to connect to Auth gRPC: %v", err)
}
log.Printf("Using gRPC auth client, connecting to %s", grpcAddr)
auth = c
} else {
authURL := os.Getenv("AUTH_BASE_URL")
if authURL == "" {
authURL = "http://localhost:8081"
}
log.Printf("Using HTTP auth client, connecting to %s", authURL)
auth = authclient.New(authURL)
}

svc := service.New()
handler := httphandler.NewWithAuth(svc, auth)

var mux http.Handler = handler.Routes()
mux = middleware.Logging(mux)
mux = middleware.RequestID(mux)

log.Printf("Tasks service starting on :%s (mode: %s)", port, authMode)
if err := http.ListenAndServe(":"+port, mux); err != nil {
log.Fatalf("Tasks service failed: %v", err)
}
}

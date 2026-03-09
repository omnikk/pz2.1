package main

import (
	"log"
	"net/http"
	"os"

	httphandler "github.com/omnik/tech-ip-sem2/services/auth/internal/http"
	"github.com/omnik/tech-ip-sem2/services/auth/internal/service"
	"github.com/omnik/tech-ip-sem2/shared/middleware"
)

func main() {
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8081"
	}

	svc := service.New()
	handler := httphandler.New(svc)

	var mux http.Handler = handler.Routes()
	mux = middleware.Logging(mux)
	mux = middleware.RequestID(mux)

	log.Printf("Auth service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Auth service failed: %v", err)
	}
}

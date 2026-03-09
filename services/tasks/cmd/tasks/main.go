package main

import (
	"log"
	"net/http"
	"os"

	"github.com/omnik/tech-ip-sem2/services/tasks/internal/client/authclient"
	httphandler "github.com/omnik/tech-ip-sem2/services/tasks/internal/http"
	"github.com/omnik/tech-ip-sem2/services/tasks/internal/service"
	"github.com/omnik/tech-ip-sem2/shared/middleware"
)

func main() {
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	authURL := os.Getenv("AUTH_BASE_URL")
	if authURL == "" {
		authURL = "http://localhost:8081"
	}

	svc := service.New()
	authClient := authclient.New(authURL)
	handler := httphandler.New(svc, authClient)

	var mux http.Handler = handler.Routes()
	mux = middleware.Logging(mux)
	mux = middleware.RequestID(mux)

	log.Printf("Tasks service starting on :%s (auth: %s)", port, authURL)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Tasks service failed: %v", err)
	}
}

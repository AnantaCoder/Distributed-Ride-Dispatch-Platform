package main

import (
	"log"
	"net/http"
	"os"
	

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/gateway"
	myMiddleware "github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// 1. Connect to Redis (for idempotency)
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr()})

	//now connecting to temp	

	// 2. Connect to Temporal (to start workflows)
	tc, err := client.Dial(client.Options{HostPort: cfg.Temporal.Addr()})
	if err != nil {
		log.Fatalln("Failed to connect to Temporal:", err)
	}
	defer tc.Close()

	// 3. Set up the Chi Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	
	// Apply your Idempotency Middleware!
	r.Use(myMiddleware.IdempotencyMiddleware(rdb))

	// 4. Register Routes
	h := gateway.NewGatewayHandlers(tc, rdb)
	r.Post("/ride", h.HandleRequestRide)
	r.Post("/driver/location", h.HandleUpdateLocation)
	r.Post("/ride/{id}/accept", h.HandleAcceptRide)

	// 5. Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Gateway full address is http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

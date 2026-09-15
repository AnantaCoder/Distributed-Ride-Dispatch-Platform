// 1. load config
// 2. initialize the infra connections (only Redis for Pricing)
// 3. dependancy injections
// 4. setup the router
// 5. asyncronous server starter
// 6. shutdown

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/pricing"
)

func main() {
	// =========================================================================
	// 1. Load config
	log.Println("Loading configuration...")
	cfg := config.LoadConfig()

	// =========================================================================
	// 2. Initialize the infra connections
	ctx := context.Background()

	// 2a. Redis (Pricing only needs Redis for dynamic surge multipliers)
	log.Println("Connecting to Redis...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping redis: %v", err)
	}
	log.Println("Redis connected successfully.")

	// =========================================================================
	// 3. Dependency injections
	log.Println("Initializing service layers...")
	svc := pricing.NewService(redisClient)
	handler := pricing.NewHandler(svc)

	// =========================================================================
	// 4. Setup the router
	log.Println("Setting up routes...")
	r := chi.NewRouter()

	// Add standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Mount the pricing REST endpoints
	r.Get("/pricing/estimate", handler.HandleEstimatePrice)

	// =========================================================================
	// 5. Asynchronous server starter
	port := cfg.Services.PricingServicePort
	addr := fmt.Sprintf(":%d", port)
	
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting Pricing Service on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", addr, err)
		}
	}()

	// =========================================================================
	// 6. Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	
	log.Println("Server is shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}

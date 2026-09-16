package main

import (
	"log"
	
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/driver"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/pricing"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/trip"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/workflow"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// 1. Create a Temporal Client
	temporalClient , err := client.Dial(client.Options{HostPort: cfg.Temporal.Addr()})
	if err != nil {
		log.Fatalln("Failed to connect to Temporal", err)
	}
	defer temporalClient.Close()

	// 2. Create a New Worker
	// Hint: Use worker.New(c, "ride-task-queue", worker.Options{})
	w := worker.New(temporalClient, "ride-task-queue", worker.Options{})
	if w == nil {
		log.Fatalln("Failed to create worker")
	}
	defer w.Stop()


	// 3. Register the Workflow and Activities
	w.RegisterWorkflow(workflow.RideLifecycleWorkflow)
	
	// --- BOILERPLATE FOR DEPENDENCY INJECTION ---
	ctx := context.Background()

	// Connect to Postgres
	pgPool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pgPool.Close()

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr()})
	defer redisClient.Close()

	// Initialize Repositories
	tripRepo := trip.NewRepository(pgPool)
	driverRepo := driver.NewRepository(pgPool)

	// Initialize Services
	pricingSvc := pricing.NewService(redisClient)
	tripSvc := trip.NewService(tripRepo, pricingSvc)
	driverSvc := driver.NewService(driverRepo, redisClient, cfg.Matching)

	// Inject services into Activities
	a := workflow.NewRideActivities(tripSvc, driverSvc, pricingSvc)
	// ---------------------------------------------

	w.RegisterActivity(a.EstimatePriceActivity)
	w.RegisterActivity(a.FindAndAssignDriverActivity)
	w.RegisterActivity(a.UpdateTripStatusActivity)

	// 4. Start the Worker!
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Worker failed to start:", err)
	}
}

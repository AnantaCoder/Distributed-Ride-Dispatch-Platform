package main

import (
	"log"
	
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
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
	
	// Create an instance of our activities
	a := &workflow.RideActivities{}
	w.RegisterActivity(a.EstimatePriceActivity)
	w.RegisterActivity(a.FindAndAssignDriverActivity)
	w.RegisterActivity(a.UpdateTripStatusActivity)

	// 4. Start the Worker!
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Worker failed to start:", err)
	}
}

package main

import (
	"log"
	
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/workflow"
)

func main() {
	// 1. Create a Temporal Client
	// Hint: Use client.Dial(client.Options{HostPort: "localhost:7233"})

	temporalClient , err := client.Dial(client.Options{HostPort:"127.0.0.1:7233"})
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

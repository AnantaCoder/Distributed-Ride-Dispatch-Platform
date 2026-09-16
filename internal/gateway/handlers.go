package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/workflow"
)

type GatewayHandlers struct {
	TemporalClient client.Client
	RedisClient    *redis.Client
}

func NewGatewayHandlers(tc client.Client, rc *redis.Client) *GatewayHandlers {
	return &GatewayHandlers{
		TemporalClient: tc,
		RedisClient:    rc,
	}
}

// POST /ride
func (h *GatewayHandlers) HandleRequestRide(w http.ResponseWriter, r *http.Request) {
	var req workflow.RideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	options := client.StartWorkflowOptions{
		ID:        "ride-" + req.RideID,
		TaskQueue: "ride-task-queue",
	}

	we, err := h.TemporalClient.ExecuteWorkflow(context.Background(), options, workflow.RideLifecycleWorkflow, req)
	if err != nil {
		http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"workflow_id": we.GetID()})
}

// POST /driver/location
func (h *GatewayHandlers) HandleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	// TODO: Forward location data to the Driver Service 
	w.WriteHeader(http.StatusOK)
}

// POST /ride/{id}/accept
func (h *GatewayHandlers) HandleAcceptRide(w http.ResponseWriter, r *http.Request) {
	rideID := chi.URLParam(r, "id")
	
	// Send a signal to the running workflow that the driver accepted
	err := h.TemporalClient.SignalWorkflow(context.Background(), "ride-"+rideID, "", "DriverAcceptedSignal", nil)
	if err != nil {
		http.Error(w, "Failed to signal workflow", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

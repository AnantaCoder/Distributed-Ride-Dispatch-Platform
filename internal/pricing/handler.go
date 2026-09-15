package pricing

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Handler handles HTTP requests for the pricing service.
type Handler struct {
	svc Service
}

// NewHandler creates a new HTTP handler for pricing.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// HandleEstimatePrice is an HTTP endpoint to get a price estimate before requesting a ride.
func (h * Handler ) HandleEstimatePrice(w http.ResponseWriter , r *http.Request) {
	query := r.URL.Query()

	pickupLatStr := query.Get("pickup_lat")
	pickupLngStr := query.Get("pickup_lng")
	dropoffLatStr := query.Get("dropoff_lat")
	dropoffLngStr := query.Get("dropoff_lng")

	if pickupLatStr == "" || pickupLngStr == "" || dropoffLatStr == "" || dropoffLngStr == "" {
		http.Error(w, "missing required query parameters", http.StatusBadRequest)
		return
	}

	pickupLat, err := strconv.ParseFloat(pickupLatStr, 64)
	if err != nil {
		http.Error(w, "invalid pickup latitude", http.StatusBadRequest)
		return
	}

	pickupLng, err := strconv.ParseFloat(pickupLngStr, 64)
	if err != nil {
		http.Error(w, "invalid pickup longitude", http.StatusBadRequest)
		return
	}

	dropoffLat, err := strconv.ParseFloat(dropoffLatStr, 64)
	if err != nil {
		http.Error(w, "invalid dropoff latitude", http.StatusBadRequest)
		return
	}

	dropoffLng, err := strconv.ParseFloat(dropoffLngStr, 64)
	if err != nil {
		http.Error(w, "invalid dropoff longitude", http.StatusBadRequest)
		return
	}

	estimate, err := h.svc.EstimatePrice(r.Context(), pickupLat, pickupLng, dropoffLat, dropoffLng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Price    float64 `json:"price"`
		Currency string  `json:"currency"`
	}{
		Price:    float64(estimate),
		Currency: CurrencyCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}


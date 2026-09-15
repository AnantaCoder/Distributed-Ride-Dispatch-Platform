package driver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/drivers/{id}/location", h.updateLocation)
	r.Post("/drivers/{id}/availability", h.setAvailability)
	r.Get("/drivers/nearby", h.getNearbyDrivers)
}

func (h *Handler) updateLocation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	driverID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid driver id", http.StatusBadRequest)
		return
	}

	var req struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.UpdateLocation(r.Context(), driverID, req.Lat, req.Lng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) setAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	driverID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid driver id", http.StatusBadRequest)
		return
	}

	var req struct {
		IsAvailable bool `json:"is_available"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.SetAvailability(r.Context(), driverID, req.IsAvailable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getNearbyDrivers(w http.ResponseWriter, r *http.Request) {
	// 1. Read query parameters from URL (e.g. /drivers/nearby?lat=37.77&lng=-122.41&radius=5000)
	// Note: We use r.URL.Query().Get() instead of chi.URLParam() because these are query params (?key=val), not path params (/{id})
	query := r.URL.Query()
	latStr := query.Get("lat")
	lngStr := query.Get("lng")
	radiusStr := query.Get("radius")

	// 2. Validate that all required parameters are present
	if latStr == "" || lngStr == "" || radiusStr == "" {
		http.Error(w, "missing required query params: lat, lng, radius", http.StatusBadRequest)
		return
	}

	// 3. Parse and convert string parameters to float64
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "invalid latitude", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		http.Error(w, "invalid longitude", http.StatusBadRequest)
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		http.Error(w, "invalid radius", http.StatusBadRequest)
		return
	}

	// 4. Call the service layer using request context (r.Context())
	drivers, err := h.service.GetNearbyDrivers(r.Context(), lat, lng, radius)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Ensure we return an empty JSON array `[]` instead of `null` if no drivers are found
	if drivers == nil {
		drivers = []Driver{}
	}

	// 6. Set response header to JSON and encode the list of drivers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(drivers); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

package workflow

import (
	"context"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/driver"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/pricing"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/trip"
)

type RideActivities struct {
	TripService    trip.Service
	DriverService  driver.Service
	PricingService pricing.Service
}

func NewRideActivities(ts trip.Service, ds driver.Service, ps pricing.Service) *RideActivities {
	return &RideActivities{
		TripService:    ts,
		DriverService:  ds,
		PricingService: ps,
	}
}

func (a *RideActivities) EstimatePriceActivity(ctx context.Context, rideID string) (int, error) {
	// Stub: Will call PricingService
	return 1500, nil
}

func (a *RideActivities) FindAndAssignDriverActivity(ctx context.Context, rideID string) (string, error) {
	// Stub: Will call DriverService matcher
	return "driver-123", nil
}

func (a *RideActivities) UpdateTripStatusActivity(ctx context.Context, rideID string, status string) error {
	// Stub: Will call TripService database update
	return nil
}
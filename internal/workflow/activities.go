package workflow

import (
	"context"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/driver"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/pricing"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/trip"
	"github.com/google/uuid"
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

func (a *RideActivities) EstimatePriceActivity(ctx context.Context, req RideRequest) (int, error) {
	// Removed DB trip fetch since Gateway doesn't create the trip yet
	
	price, err := a.PricingService.EstimatePrice(ctx, req.Lat, req.Lng, req.Lat+0.05, req.Lng+0.05)
	if err != nil {
		return 0, err
	}
	
	return int(price), nil
}

func (a *RideActivities) FindAndAssignDriverActivity(ctx context.Context, req RideRequest) (string, error) {
	bestDriver, err := a.DriverService.FindBestDriver(ctx, req.Lat, req.Lng)
	if err != nil {
		return "", err
	}
	
	return bestDriver.ID.String(), nil
}

func (a *RideActivities) UpdateTripStatusActivity(ctx context.Context, rideID string, status string) error {
	_, err := a.TripService.UpdateTrip(ctx, uuid.MustParse(rideID), status)
	return err
}
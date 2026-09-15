package trip

import (
	"context"
	"errors"
	"time"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/pricing"
	"github.com/google/uuid"
)

// Service defines the business logic operations for trips.
type Service interface {
	GetTrip(ctx context.Context, id uuid.UUID) (*Trip, error)
	RequestRide(ctx context.Context, passengerID uuid.UUID, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (*Trip, error)
	CancelTrip(ctx context.Context, id uuid.UUID) (*Trip, error)
	UpdateTrip(ctx context.Context, id uuid.UUID, newStatus string) (*Trip , error)
}

// tripService is the concrete implementation of the Service interface.
type tripService struct {
	repo    Repository
	pricing pricing.Service
}

// NewService creates a new instance of the trip Service with its dependencies.
func NewService(repo Repository, pricing pricing.Service) Service {
	return &tripService{
		repo:    repo,
		pricing: pricing,
	}
}

// GetTrip retrieves a trip by its unique identifier.
func (s *tripService) GetTrip(ctx context.Context, id uuid.UUID) (*Trip, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid trip id")
	}
	trip, err := s.repo.GetTripByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, errors.New("trip not found")
	}
	return trip, nil
}

// calculateEstimatePrice delegates fare calculation to the pricing service.
func (s *tripService) calculateEstimatePrice(
	ctx context.Context,
	pickupLat, pickupLng, dropoffLat, dropoffLng float64,
) (int64, error) {
	if s.pricing == nil {
		return 0, errors.New("pricing service not configured")
	}
	return s.pricing.EstimatePrice(ctx, pickupLat, pickupLng, dropoffLat, dropoffLng)
}
// RequestRide creates and dispatches a new trip request for a passenger.
func (s *tripService) RequestRide(
	ctx context.Context,
	passengerID uuid.UUID,
	pickupLat float64,
	pickupLng float64,
	dropoffLat float64,
	dropoffLng float64,
) (*Trip, error) {
	//validation -> creating trip -> estimating price -> updating price -> persisting -> returning trip
	if passengerID == uuid.Nil {
		return nil, errors.New("invalid passenger id")
	}
	if pickupLat > 90 || pickupLat < -90 || pickupLng > 180 || pickupLng < -180 {
		return nil, errors.New("invalid pickup coordinates")
	}
	if dropoffLat > 90 || dropoffLat < -90 || dropoffLng > 180 || dropoffLng < -180 {
		return nil, errors.New("invalid dropoff coordinates")
	}

	id := uuid.New() // new 

	newTrip := &Trip{
		ID : id,
		PassengerID : passengerID,
		PickupLat : pickupLat,
		PickupLng : pickupLng,
		DropoffLat : dropoffLat,
		DropoffLng : dropoffLng,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status : "REQUESTED",
	}
	//estimate price
	price ,err := s.calculateEstimatePrice(ctx , pickupLat,pickupLng,dropoffLat,dropoffLng)
	if err == nil {
		newTrip.EstimatedPriceCents = &price
	}
	err = s.repo.CreateTrip(ctx, newTrip)
	if err != nil {
		return nil, err
	}

	return newTrip, nil
}

// CancelTrip cancels an ongoing or requested trip.
func (s *tripService) CancelTrip(ctx context.Context, id uuid.UUID) (*Trip, error) {
	//id validation -> getting trip details -> status check -> updating status-> return updated trip
	if id == uuid.Nil{
		return nil, errors.New("invalid trip id")
	}

	trip, err := s.repo.GetTripByID(ctx,id)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, errors.New("trip not found")
	}
	if trip.Status == "COMPLETED" || trip.Status == "CANCELLED" {
		return nil, errors.New("trip cannot be cancelled")
	}

    err = s.repo.UpdateTripStatus(ctx,id,"CANCELLED")
	if err != nil {
		return nil, err
	}
	trip.Status = "CANCELLED"
	return trip, nil
}

// UpdateTrip updates the status of an existing trip.
func (s *tripService) UpdateTrip(ctx context.Context, id uuid.UUID, newStatus string) (*Trip, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid trip id")
	}
	trip, err := s.repo.GetTripByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, errors.New("trip not found")
	}
	err = s.repo.UpdateTripStatus(ctx, id, newStatus)
	if err != nil {
		return nil, err
	}
	trip.Status = newStatus
	return trip, nil
}

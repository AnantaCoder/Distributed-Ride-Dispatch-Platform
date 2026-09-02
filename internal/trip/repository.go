package trip

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrTripNotFound is returned when a trip is not found in the database.
var ErrTripNotFound = errors.New("trip not found")

// Trip represents the domain model for a ride.
// Note: This matches the schema in your implementation plan.
type Trip struct {
	ID                  uuid.UUID
	PassengerID         uuid.UUID
	DriverID            *uuid.UUID // Pointer because it can be null before assignment
	PickupLat           float64
	PickupLng           float64
	DropoffLat          float64
	DropoffLng          float64
	Status              string
	EstimatedPriceCents *int64
	FinalPriceCents     *int64
	IdempotencyKey      *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// Repository defines the data access methods for trips.
type Repository interface {
	CreateTrip(ctx context.Context, trip *Trip) error
	GetTripByID(ctx context.Context, id uuid.UUID) (*Trip, error)
	UpdateTripStatus(ctx context.Context, id uuid.UUID, status string) error
}

// tripRepository is the concrete implementation of Repository using PostgreSQL.
type tripRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgreSQL-backed trip repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &tripRepository{
		pool: pool,
	}
}

// CreateTrip inserts a new trip into the database.
func (r *tripRepository) CreateTrip(ctx context.Context, trip *Trip) error {

	query := `INSERT INTO trips (passenger_id,pickup_lat,pickup_lng,dropoff_lat,dropoff_lng,status,idempotency_key,estimated_price_cents) values ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at,updated_at`

	err := r.pool.QueryRow(ctx, query, 
    trip.PassengerID, trip.PickupLat, trip.PickupLng, trip.DropoffLat, trip.DropoffLng, trip.Status, trip.IdempotencyKey, trip.EstimatedPriceCents,
    ).Scan(&trip.ID, &trip.CreatedAt, &trip.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetTripByID fetches a trip from the database.
func (r *tripRepository) GetTripByID(ctx context.Context, id uuid.UUID) (*Trip, error) {
	
	query := `Select id, passenger_id, driver_id, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng, status, estimated_price_cents, final_price_cents, idempotency_key, created_at, updated_at from trips WHERE id = $1`
	trip := &Trip{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&trip.ID, &trip.PassengerID, &trip.DriverID, &trip.PickupLat, &trip.PickupLng, &trip.DropoffLat, &trip.DropoffLng, &trip.Status, &trip.EstimatedPriceCents, &trip.FinalPriceCents, &trip.IdempotencyKey, &trip.CreatedAt, &trip.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrTripNotFound
	}
	
	if err != nil {
		return nil, err
	}
	return trip, nil
}

// UpdateTripStatus updates the status of an existing trip.
func (r *tripRepository) UpdateTripStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE trips SET status = $1, updated_at = NOW() WHERE id = $2 `

	commandTag, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrTripNotFound
	}

	return nil
}

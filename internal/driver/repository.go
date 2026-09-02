package driver

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)




var ErrDriverNotFound = errors.New("driver not found")

type Driver struct{
	ID          uuid.UUID
	DriverID    uuid.UUID
	CarPlateNumber string
	CarModel       string
	CarColor       string
	IsAvailable    bool
	CurrentLat     float64
	CurrentLng     float64
	Rating         int
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}


// repository interface for database operations on driver table 
type Repository interface {
	CreateDriver(ctx context.Context , driver *Driver) error
	GetDriverByID(ctx context.Context , id uuid.UUID) (*Driver, error)
	UpdateDriverStatus(ctx context.Context , id uuid.UUID, status string) error
}
// driver repository implementation 
type driverRepository struct {
	pool *pgxpool.Pool
}
// new repository for postgres based trip 
func NewRepository(pool *pgxpool.Pool) Repository {
	return &driverRepository{
		pool: pool,
	}
}
// create driver 
func (r *driverRepository)CreateDriver(ctx context.Context , driver *Driver) error {
	query := `INSERT INTO drivers (driver_id,car_plate_number,car_model,car_color,is_available,current_lat,current_lng,rating,version) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_,err := r.pool.Exec(ctx, query, driver.DriverID, driver.CarPlateNumber, driver.CarModel, driver.CarColor, driver.IsAvailable, driver.CurrentLat, driver.CurrentLng, driver.Rating, driver.Version)
	if err != nil {
		return err
	}
	return nil
}
//get driver by id 
func (r *driverRepository) GetDriverByID(ctx context.Context , id uuid.UUID) (*Driver, error) {
	query := `SELECT id,driver_id,car_plate_number,car_model,car_color,is_available,current_lat,current_lng,rating,version,created_at,updated_at FROM drivers WHERE id = $1`
	driver := &Driver{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&driver.ID, &driver.DriverID, &driver.CarPlateNumber, &driver.CarModel, &driver.CarColor, &driver.IsAvailable, &driver.CurrentLat, &driver.CurrentLng, &driver.Rating, &driver.Version, &driver.CreatedAt, &driver.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrDriverNotFound
	}
	if err != nil {
		return nil, err
	}
	return driver, nil
}
// update driver status 
func (r *driverRepository) UpdateDriverStatus(ctx context.Context , id uuid.UUID, status string) error {
	query := `UPDATE drivers SET is_available = $1, updated_at = NOW() WHERE id = $2`
	commandTag, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrDriverNotFound
	}
	return nil
}



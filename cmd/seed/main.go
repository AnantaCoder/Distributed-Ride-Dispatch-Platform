package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/driver"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	// 1. Connect to Postgres
	pgPool, err := pgxpool.New(ctx, cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("Failed to create postgres pool: %v", err)
	}
	defer pgPool.Close()

	// 2. Connect to Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr()})
	defer redisClient.Close()

	// 3. Create a fake driver
	driverID := uuid.New()
	
	fakeDriver := &driver.Driver{
		DriverID:       driverID,
		CarPlateNumber: "ABC-1234",
		CarModel:       "Toyota Prius",
		CarColor:       "White",
		IsAvailable:    true,
		CurrentLat:     37.77,   // Exactly where our trip requests are happening!
		CurrentLng:     -122.41,
		Rating:         5,
		Version:        1,
	}

	// 4. Insert into Postgres using the repository
	repo := driver.NewRepository(pgPool)
	
	err = repo.CreateDriver(ctx, fakeDriver)
	if err != nil {
		log.Printf("Failed to insert driver to postgres, falling back to raw SQL... Error: %v\n", err)
		var realID uuid.UUID
		err = pgPool.QueryRow(ctx, "INSERT INTO drivers (driver_id, car_plate_number, car_model, car_color, is_available, current_lat, current_lng, rating, version) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id", 
			fakeDriver.DriverID, fakeDriver.CarPlateNumber, fakeDriver.CarModel, fakeDriver.CarColor, fakeDriver.IsAvailable, fakeDriver.CurrentLat, fakeDriver.CurrentLng, fakeDriver.Rating, fakeDriver.Version).Scan(&realID)
		if err != nil {
			log.Fatalf("Fatal DB error: %v", err)
		}
		driverID = realID
	} else {
		// If CreateDriver worked, we fetch the DB generated UUID back
		var realID uuid.UUID
		err = pgPool.QueryRow(ctx, "SELECT id FROM drivers WHERE driver_id = $1 LIMIT 1", driverID).Scan(&realID)
		if err != nil {
			log.Fatalf("Could not fetch inserted driver ID: %v", err)
		}
		driverID = realID
	}

	// 5. Add to Redis Geo
	err = redisClient.GeoAdd(ctx, "drivers:locations", &redis.GeoLocation{
		Name:      driverID.String(),
		Longitude: fakeDriver.CurrentLng,
		Latitude:  fakeDriver.CurrentLat,
	}).Err()
	if err != nil {
		log.Fatalf("Failed to add to Redis: %v", err)
	}

	fmt.Printf("✅ Success! Seeded Driver '%s' into PostgreSQL and Redis!\n", driverID.String())
}

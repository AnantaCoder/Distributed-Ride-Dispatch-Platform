package pricing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// Currency and standard Indian ride-hailing rates
// All monetary values are calculated in the smallest unit: Paise (1 Rupee = 100 Paise).
const (
	CurrencyCode = "INR"
	// Base Fare: ₹50.00 (5,000 paise) includes the first 1.5 km
	DefaultBaseFarePaise int64 = 5000
	// Base Distance included in base fare (in kilometers)
	BaseDistanceKm float64 = 1.5
	// Per Kilometer Rate after base distance: ₹15.00/km (1,500 paise/km)
	DefaultCostPerKmPaise int64 = 1500
	// Minimum Fare: ₹60.00 (6,000 paise)
	DefaultMinimumFarePaise int64 = 6000
	// GST rate for ride aggregators in India: 5% (2.5% CGST + 2.5% SGST)
	GSTRate float64 = 0.05
	// Night charge multiplier (25% extra between 11:00 PM and 5:00 AM IST)
	NightChargeMultiplier float64 = 1.25
	EarthRadiusKm float64 = 6371.0
)

// IST location (Indian Standard Time, UTC +5:30)
var istLocation = time.FixedZone("IST", 5*3600+30*60)

// Service defines the pricing business logic operations.
type Service interface {
	// EstimatePrice calculates an estimated fare in Paise (INR) between pickup and dropoff points.
	EstimatePrice(ctx context.Context, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (int64, error)
	// GetSurgeMultiplier calculates dynamic pricing multiplier based on regional demand/supply.
	GetSurgeMultiplier(ctx context.Context, lat, lng float64) (float64, error)
}

// pricingService is the concrete implementation of Service.
type pricingService struct {
	redis *redis.Client // optional: used for dynamic surge calculation (supply vs demand)
}

// NewService creates a new Pricing Service instance.
func NewService(redisClient *redis.Client) Service {
	return &pricingService{
		redis: redisClient,
	}
}

// EstimatePrice calculates the estimated ride fare in Indian Paise (₹ / 100) using:
// 1. Haversine distance in km
// 2. Base fare (first 1.5 km) + Per-km rate for extra distance
// 3. Night-time multiplier (11 PM - 5 AM IST)
// 4. Dynamic demand/supply surge multiplier
// 5. 5% GST (CGST + SGST)
// 6. Minimum fare enforcement (₹60.00)
func (s *pricingService) EstimatePrice(
	ctx context.Context,
	pickupLat float64,
	pickupLng float64,
	dropoffLat float64,
	dropoffLng float64,
) (int64, error) {
	// 1. Validate coordinates
	if pickupLat < -90 || pickupLat > 90 || pickupLng < -180 || pickupLng > 180 {
		return 0, errors.New("invalid pickup coordinates")
	}
	if dropoffLat < -90 || dropoffLat > 90 || dropoffLng < -180 || dropoffLng > 180 {
		return 0, errors.New("invalid dropoff coordinates")
	}

	// 2. Calculate distance in kilometers
	distanceKm := calculateDistanceKm(pickupLat, pickupLng, dropoffLat, dropoffLng)

	// 3. Base fare + Distance fare
	farePaise := float64(DefaultBaseFarePaise)
	if distanceKm > BaseDistanceKm {
		extraKm := distanceKm - BaseDistanceKm
		farePaise += extraKm * float64(DefaultCostPerKmPaise)
	}

	// 4. Night-time multiplier (11:00 PM to 05:00 AM IST)
	nowIST := time.Now().In(istLocation)
	hour := nowIST.Hour()
	if hour >= 23 || hour < 5 {
		farePaise *= NightChargeMultiplier
	}

	// 5. Dynamic surge multiplier (traffic / high demand)
	surge, err := s.GetSurgeMultiplier(ctx, pickupLat, pickupLng)
	if err == nil && surge > 1.0 {
		farePaise *= surge
	}

	// 6. Add 5% GST (Goods and Services Tax)
	fareWithTax := farePaise * (1.0 + GSTRate)
	finalFarePaise := int64(math.Round(fareWithTax))

	// 7. Enforce minimum fare (₹60.00 = 6,000 paise)
	if finalFarePaise < DefaultMinimumFarePaise {
		finalFarePaise = DefaultMinimumFarePaise
	}

	return finalFarePaise, nil
}

// GetSurgeMultiplier calculates dynamic pricing based on supply/demand in India.
func (s *pricingService) GetSurgeMultiplier(ctx context.Context, lat, lng float64) (float64, error) {
	// If Redis is not connected yet, default to standard multiplier 1.0 (no surge)
	if s.redis == nil {
		return 1.0, nil
	}

	// 1. Query available drivers within 3km radius using Redis GEOSEARCH
	driverQuery := &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     3,
		RadiusUnit: "km",
	}
	drivers, err := s.redis.GeoSearch(ctx, "driver_locations", driverQuery).Result()
	if err != nil && err != redis.Nil {
		// Log error in a real system; for now, fail open and return no surge.
		return 1.0, nil 
	}
	supply := float64(len(drivers))

	// 2. Query active ride requests in the same zone
	// We assume active requests are also stored in a geospatial set called "active_requests"
	rideQuery := &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     3,
		RadiusUnit: "km",
	}
	requests, err := s.redis.GeoSearch(ctx, "active_requests", rideQuery).Result()
	if err != nil && err != redis.Nil {
		return 1.0, nil
	}
	demand := float64(len(requests))

	// 3. If demand > supply, compute surge factor: 1.0 + (demand - supply) * factor (capped at 2.5x per Indian regulations)
	surge := 1.0
	if demand > supply && supply > 0 {
		// Increase by 0.2 for every extra ride request that exceeds supply
		surge += (demand - supply) * 0.2
	} else if demand > 0 && supply == 0 {
		// Max surge when there is demand but zero supply
		surge = 2.5
	}

	// Cap the surge at 2.5x maximum
	if surge > 2.5 {
		surge = 2.5
	}

	return surge, nil
}

// FormatINR converts paise into formatted Indian Rupee string (e.g., 15450 paise -> "₹154.50").
func FormatINR(paise int64) string {
	rupees := float64(paise) / 100.0
	return fmt.Sprintf("₹%.2f", rupees)
}

// PaiseToRupees converts paise to standard Rupee value.
func PaiseToRupees(paise int64) float64 {
	return float64(paise) / 100.0
}

// calculateDistanceKm computes the great-circle distance between two points using the Haversine formula.
func calculateDistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLng := (lng2 - lng1) * (math.Pi / 180.0)

	radLat1 := lat1 * (math.Pi / 180.0)
	radLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLng/2)*math.Sin(dLng/2)*math.Cos(radLat1)*math.Cos(radLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}
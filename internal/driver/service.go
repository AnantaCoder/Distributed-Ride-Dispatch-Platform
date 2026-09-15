package driver

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// setup an interface and then concrect struct that holds dependencies , and a  constructor
//Service defines the business logic methods for driver operations.

//operations the driver service must provide 
type Service interface{
UpdateLocation(ctx context.Context , driverID uuid.UUID , lat float64 , lng float64 ) error
SetAvailability(ctx context.Context , driverID uuid.UUID , isAvailable bool) error
GetNearbyDrivers(ctx context.Context , lat float64 , lng float64 , radiusInMeters float64 ) ([]Driver , error)
}

//driver service is a concreate(actual) implimentation of the service 
type driverService struct{
	repo Repository
	redis *redis.Client
}


//new service now create a driver service with its dependencies 
//constructor 
func NewService(repo Repository , redisClient *redis.Client ) Service{
	return &driverService{
		repo : repo,
		redis : redisClient,
	}
}

func (s *driverService) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat float64, lng float64) error {
	//validate 
	if driverID == uuid.Nil {
		return errors.New("invalid driver id")
	}

	//update location 
	driver,err := s.repo.GetDriverByID(ctx , driverID)
	if err != nil {
		return err
	}

	if driver == nil {
		return errors.New("driver not found")
	}

	//use redis geo adding/updating the driver's location in Redis's geospatial index
	err = s.redis.GeoAdd(ctx, "drivers:locations", &redis.GeoLocation{
		Name:      driverID.String(),
		Longitude: lng,
		Latitude:  lat,
	}).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *driverService) SetAvailability(ctx context.Context, driverID uuid.UUID, isAvailable bool) error {
	
	if driverID == uuid.Nil {
		return errors.New("invalid driver id")
	}

	//fetch the driver profile  
	driver, err := s.repo.GetDriverByID(ctx , driverID)
	if err != nil {
		return err
	}
	if driver == nil {
		return errors.New("driver not found")
	}

	//initially assuming the avilibility to false 
	statusStr := "false"
	if isAvailable {
		statusStr = "true"
	}

	//updating the status in the db of that particular driver 
	err = s.repo.UpdateDriverStatus(ctx, driverID, statusStr)
	if err != nil {
		return err
	}

	// this is the redis update logic 
	//if the driver is available then add his location to the redis geo 
	//else remove it from the geo so that it is not considered in the nearby drivers query 
	if isAvailable {
		return s.redis.GeoAdd(ctx, "drivers:locations", &redis.GeoLocation{
			Name:      driverID.String(),
			Longitude: driver.CurrentLng,
			Latitude:  driver.CurrentLat,
		}).Err()
	} else {
		return s.redis.ZRem(ctx, "drivers:locations", driverID.String()).Err()
	}
	
	
}

//this function is to get the nearby drivers by querying the redis and then finding their info from the database 
func (s *driverService) GetNearbyDrivers(ctx context.Context, lat float64, lng float64, radiusInMeters float64) ([]Driver, error) {
	// query redis for drivers within radius
	res, err := s.redis.GeoSearchLocation(ctx, "drivers:locations", &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radiusInMeters,
			RadiusUnit: "m", //meters
		},
		WithCoord: true,// gives drivers latitude and longitude 
		WithDist:  true, //distance from the search location 
	}).Result()

	if err != nil && err != redis.Nil {
		return nil, err
	}

	var drivers []Driver
	//iterating throug the locations in redis 
	for _, loc := range res {
		//Parsing the id 
		driverID, err := uuid.Parse(loc.Name)
		if err != nil {
			continue //skipping the driver if the id is invalid 
		}

		driver, err := s.repo.GetDriverByID(ctx, driverID) //fetcht the driver from DB
		if err != nil {
			continue //skipping the that driver if it throws error in db
		}

		// use the real-time location from Redis 
		driver.CurrentLat = loc.Latitude
		driver.CurrentLng = loc.Longitude

		drivers = append(drivers, *driver)
	}

	return drivers, nil
}

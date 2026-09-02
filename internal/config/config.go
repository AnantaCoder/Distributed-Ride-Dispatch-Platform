package config

// This is the "no external dependency" pattern — instead of pulling in Viper or envconfig,
// you use stdlib only. Interviewers love this because it shows you understand that a config
// loader is just os.Getenv + type conversion + defaults.

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the platform.
type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
	Temporal TemporalConfig
	Services ServicesConfig
	Matching MatchingConfig
	Workflow WorkflowConfig
}

// PostgresConfig holds PostgreSQL connection settings.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DB, c.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Redis address in host:port format.
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// TemporalConfig holds Temporal connection settings.
type TemporalConfig struct {
	Host      string
	Port      int
	Namespace string
	TaskQueue string
}

// Addr returns the Temporal frontend address.
func (c TemporalConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ServicesConfig holds ports for each microservice.
type ServicesConfig struct {
	TripServicePort    int
	DriverServicePort  int
	PricingServicePort int
	GatewayPort        int
}

// MatchingConfig holds driver-matching algorithm parameters.
type MatchingConfig struct {
	MaxRadiusKm          float64
	MaxCandidates        int
	DistanceWeight       float64
	RatingWeight         float64
	AcceptanceRateWeight float64
	IdleTimeWeight       float64
}

// WorkflowConfig holds Temporal workflow timeout settings.
type WorkflowConfig struct {
	DriverResponseTimeout time.Duration
	MaxDriverAttempts     int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Postgres: PostgresConfig{
			Host:     envOrDefault("POSTGRES_HOST", "localhost"),
			Port:     envOrDefaultInt("POSTGRES_PORT", 5432),
			User:     envOrDefault("POSTGRES_USER", "ridedispatch"),
			Password: envOrDefault("POSTGRES_PASSWORD", "ridedispatch"),
			DB:       envOrDefault("POSTGRES_DB", "ridedispatch"),
			SSLMode:  envOrDefault("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     envOrDefault("REDIS_HOST", "localhost"),
			Port:     envOrDefaultInt("REDIS_PORT", 6379),
			Password: envOrDefault("REDIS_PASSWORD", ""),
			DB:       envOrDefaultInt("REDIS_DB", 0),
		},
		Temporal: TemporalConfig{
			Host:      envOrDefault("TEMPORAL_HOST", "localhost"),
			Port:      envOrDefaultInt("TEMPORAL_PORT", 7233),
			Namespace: envOrDefault("TEMPORAL_NAMESPACE", "default"),
			TaskQueue: envOrDefault("TEMPORAL_TASK_QUEUE", "ride-dispatch"),
		},
		Services: ServicesConfig{
			TripServicePort:    envOrDefaultInt("TRIP_SERVICE_PORT", 8081),
			DriverServicePort:  envOrDefaultInt("DRIVER_SERVICE_PORT", 8082),
			PricingServicePort: envOrDefaultInt("PRICING_SERVICE_PORT", 8083),
			GatewayPort:        envOrDefaultInt("GATEWAY_PORT", 8080),
		},
		Matching: MatchingConfig{
			MaxRadiusKm:          envOrDefaultFloat("MATCH_MAX_RADIUS_KM", 5.0),
			MaxCandidates:        envOrDefaultInt("MATCH_MAX_CANDIDATES", 10),
			DistanceWeight:       envOrDefaultFloat("MATCH_DISTANCE_WEIGHT", 0.4),
			RatingWeight:         envOrDefaultFloat("MATCH_RATING_WEIGHT", 0.25),
			AcceptanceRateWeight: envOrDefaultFloat("MATCH_ACCEPTANCE_RATE_WEIGHT", 0.2),
			IdleTimeWeight:       envOrDefaultFloat("MATCH_IDLE_TIME_WEIGHT", 0.15),
		},
		Workflow: WorkflowConfig{
			DriverResponseTimeout: time.Duration(envOrDefaultInt("DRIVER_RESPONSE_TIMEOUT_SEC", 30)) * time.Second,
			MaxDriverAttempts:     envOrDefaultInt("MAX_DRIVER_ATTEMPTS", 3),
		},
	}
}

// envOrDefault returns the value of the environment variable or the default.
func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// envOrDefaultInt returns the int value of the environment variable or the default.
func envOrDefaultInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}

// envOrDefaultFloat returns the float64 value of the environment variable or the default.
func envOrDefaultFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}
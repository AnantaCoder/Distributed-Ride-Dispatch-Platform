package driver

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
)

func TestMatcher_RankDrivers(t *testing.T) {
	// Standard matching config to test against
	cfg := config.MatchingConfig{
		MaxRadiusKm:          5.0,
		MaxCandidates:        3, // Only return top 3
		DistanceWeight:       0.4,
		RatingWeight:         0.25,
		AcceptanceRateWeight: 0.2,
		IdleTimeWeight:       0.15,
	}

	matcher := NewMatcher(cfg)

	// Driver 1: Very close, bad rating
	d1 := MatchCandidate{
		DriverID:           uuid.New(),
		Distance:           0.5, // 0.5km
		Rating:             3.0,
		RideAcceptanceRate: 0.8,
		IdleTime:           5 * time.Minute,
	}

	// Driver 2: Further away, excellent rating and long idle time (Should score high)
	d2 := MatchCandidate{
		DriverID:           uuid.New(),
		Distance:           2.5, // 2.5km
		Rating:             5.0,
		RideAcceptanceRate: 0.95,
		IdleTime:           20 * time.Minute,
	}

	// Driver 3: Middle distance, average rating, terrible acceptance rate
	d3 := MatchCandidate{
		DriverID:           uuid.New(),
		Distance:           1.5,
		Rating:             4.2,
		RideAcceptanceRate: 0.2,
		IdleTime:           10 * time.Minute,
	}

	// Driver 4: Will get truncated because MaxCandidates is 3
	d4 := MatchCandidate{
		DriverID:           uuid.New(),
		Distance:           4.0,
		Rating:             4.0,
		RideAcceptanceRate: 0.5,
		IdleTime:           1 * time.Minute,
	}

	candidates := []MatchCandidate{d1, d2, d3, d4}

	ranked := matcher.RankDrivers(candidates)

	// 1. Verify MaxCandidates truncation worked
	if len(ranked) != 3 {
		t.Errorf("Expected 3 candidates, got %d", len(ranked))
	}

	// 2. Verify sorting is descending by score
	for i := 0; i < len(ranked)-1; i++ {
		if ranked[i].Score < ranked[i+1].Score {
			t.Errorf("List is not sorted descending. ranked[%d].Score (%f) < ranked[%d].Score (%f)", 
				i, ranked[i].Score, i+1, ranked[i+1].Score)
		}
	}

	// Verify Driver 2 won
	if ranked[0].DriverID != d2.DriverID {
		t.Errorf("Expected Driver 2 (ID: %s) to win, but got %s", d2.DriverID, ranked[0].DriverID)
	}
}

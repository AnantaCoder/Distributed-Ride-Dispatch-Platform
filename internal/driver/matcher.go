package driver

import (
	"sort"
	"time"

	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
	"github.com/google/uuid"
)

// MatchCandidate represents a driver being considered for a ride assignment.
type MatchCandidate struct {
	DriverID       uuid.UUID
	Distance       float64       
	Rating         float64       
	RideAcceptanceRate float64       
	IdleTime       time.Duration 
	Score          float64       
}

// Matcher encapsulates the logic for scoring and ranking drivers.
type Matcher struct {
	cfg config.MatchingConfig
}

// NewMatcher creates a new driver matcher using the provided config.
func NewMatcher(cfg config.MatchingConfig) *Matcher {
	return &Matcher{cfg: cfg}
}

// RankDrivers calculates a score for each candidate and sorts them descending by score.
// Score formula: w1*(1-normDist) + w2*normRating + w3*acceptRate + w4*normIdleTime
func (m *Matcher) RankDrivers(candidates []MatchCandidate) []MatchCandidate {
	// 1. Find the maximum distance and maximum idle time among all candidates (for normalization).
	maxDistance := float64(0)
	maxIdleTime := time.Duration(0)

	for _, c := range candidates{
		if c.Distance > maxDistance {
				maxDistance = c.Distance
		}
		if c.IdleTime > maxIdleTime {
			maxIdleTime = c.IdleTime
		}
	}

	// to avoid division by zero if all candidates are at the same location 
	if maxDistance == 0 {
		maxDistance = 1
	}
	if maxIdleTime == 0 {
		maxIdleTime = 1
	}
	// 2. Loop through candidates and calculate their Score using the formula and weights from m.cfg.
	for i:= range candidates{
		c := candidates[i]

		normDist := c.Distance / maxDistance // normalize it 
		normRating := c.Rating / 5.0 //assuming 5 star rating system 
		normAcceptRate := c.RideAcceptanceRate
		normIdleTime := float64(c.IdleTime) / float64(maxIdleTime)

		// Calculate the final score using the config weights
		candidates[i].Score = m.cfg.DistanceWeight*(1-normDist) +
							  m.cfg.RatingWeight*normRating +
							  m.cfg.AcceptanceRateWeight*normAcceptRate +
							  m.cfg.IdleTimeWeight*normIdleTime
	}
	// 3. Sort the candidates slice so the highest score is at index 0.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// 4. If the list is longer than m.cfg.MaxCandidates, slice it to return only the top candidates.
	if len(candidates) > m.cfg.MaxCandidates {
		candidates = candidates[:m.cfg.MaxCandidates]
	}
	
	return candidates
}

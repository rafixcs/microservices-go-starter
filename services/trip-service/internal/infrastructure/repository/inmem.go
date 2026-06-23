package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"sync"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel

	mu sync.Mutex
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trips[trip.ID.Hex()] = trip
	return r.trips[trip.ID.Hex()], nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rideFares[f.ID.Hex()] = f
	return nil
}

func (r *inmemRepository) GetFare(ctx context.Context, fareID, userID string) (*domain.RideFareModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fare, exits := r.rideFares[fareID]
	if !exits {
		return nil, fmt.Errorf("fare does not exist with ID: %s", fareID)
	}

	return fare, nil
}

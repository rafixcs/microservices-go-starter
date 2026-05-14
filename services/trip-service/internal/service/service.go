package service

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewTripService(tripRepo domain.TripRepository) *service {
	return &service{
		repo: tripRepo,
	}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserId:   "",
		Status:   "",
		RideFare: fare,
	}

	return s.repo.CreateTrip(ctx, t)
}

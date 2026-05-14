package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {

	ctx := context.Background()

	inmemRepo := repository.NewInmemRepository()

	svc := service.NewTripService(inmemRepo)

	fare := &domain.RideFareModel{
		ID:                primitive.NewObjectID(),
		UserId:            "",
		PackageSlug:       "",
		TotalPriceInCents: 0.0,
	}
	t, err := svc.CreateTrip(ctx, fare)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(t)

	for {
		time.Sleep(time.Second)
	}

}

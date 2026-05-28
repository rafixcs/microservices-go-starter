package main

import (
	"context"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func handleCreateTrip(w http.ResponseWriter, r *http.Request) {
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

	response := &contracts.APIResponse{
		Data: t,
	}

	util.WriteJSON(w, http.StatusCreated, response)
}

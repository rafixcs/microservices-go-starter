package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/types"
	"ride-sharing/shared/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HttpHandler struct {
	Service domain.TripService
}

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (s *HttpHandler) HandleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		errMsg := fmt.Sprintf("failed to parse JSON data: %s", err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	t, err := s.Service.GetRoute(ctx, &reqBody.Pickup, &reqBody.Destination)
	if err != nil {
		log.Println(err)
	}

	if err := util.WriteJSON(w, http.StatusOK, t); err != nil {
		http.Error(w, "failed to parse json response", http.StatusInternalServerError)
		return
	}
}

func (s *HttpHandler) HandleCreateTrip(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	fare := &domain.RideFareModel{
		ID:                primitive.NewObjectID(),
		UserId:            "",
		PackageSlug:       "",
		TotalPriceInCents: 0.0,
	}

	t, err := s.Service.CreateTrip(ctx, fare)
	if err != nil {
		log.Println(err)
		return
	}

	response := &contracts.APIResponse{
		Data: t,
	}

	util.WriteJSON(w, http.StatusCreated, response)
}

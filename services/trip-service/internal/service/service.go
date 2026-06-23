package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripService struct {
	repo domain.TripRepository
}

func NewTripService(tripRepo domain.TripRepository) *TripService {
	return &TripService{
		repo: tripRepo,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {

	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserId:   fare.UserId,
		Status:   "pending",
		RideFare: fare,
		Driver:   domain.TripDriver{},
	}

	tripCreated, err := s.repo.CreateTrip(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("[tripService.CreateTrip] failed to create trip: %w", err)
	}
	return tripCreated, nil
}

func (s *TripService) GetAndValidateFare(ctx context.Context, fareID, userID string) (*domain.RideFareModel, error) {
	fare, err := s.repo.GetFare(ctx, fareID, userID)
	if err != nil {
		log.Printf("failed to get trip fare: %w", err)
		return nil, fmt.Errorf("failed to get trip fare: %w", err)
	}

	if fare == nil {
		log.Printf("fare not found")
		return nil, fmt.Errorf("fare not found")
	}

	if fare.UserId != userID {
		log.Printf("fare user not valid")
		return nil, fmt.Errorf("fare user not valid")
	}

	return fare, nil
}

func (s *TripService) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route from OSRM API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response: %v", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received error response from osrm api -> [%s]: %s", resp.Status, string(body))
	}

	//log.Printf("osrm response: %s", string(body))

	var routeResp tripTypes.OsrmApiResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse Osmr API reponse: %v", err)
	}

	log.Printf("route response: %v", routeResp.Routes[0].Geometry.Coordinates[0][0])

	return &routeResp, nil
}

func (s *TripService) EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	baseFares := getBaseFares()

	estimatedFares := make([]*domain.RideFareModel, len(baseFares))
	for i, f := range baseFares {
		estimatedFares[i] = s.estimateFareRoute(f, route)
	}

	return estimatedFares, nil
}

func (s *TripService) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &domain.RideFareModel{
			ID:                id,
			UserId:            userID,
			PackageSlug:       f.PackageSlug,
			TotalPriceInCents: f.TotalPriceInCents,
			Route:             route,
		}

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %w", err)
		}

		fares[i] = fare
	}

	return fares, nil
}

func (s *TripService) estimateFareRoute(fare *domain.RideFareModel, route *tripTypes.OsrmApiResponse) *domain.RideFareModel {
	princingCfg := tripTypes.DefaultPricingConfig()
	carPackagePrice := fare.TotalPriceInCents

	distanceKm := route.Routes[0].Distance / 1000
	durationInMin := route.Routes[0].Duration / 60

	log.Printf("distanceKm %f --- durationInMin %f", distanceKm, durationInMin)

	distanceFare := distanceKm * princingCfg.PricePerUnitOfDistance
	timeFare := durationInMin * princingCfg.PricingPerMiniute
	totalPrice := carPackagePrice + distanceFare*timeFare

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug:       fare.PackageSlug,
	}
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}

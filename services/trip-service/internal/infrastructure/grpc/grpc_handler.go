package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/service"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service *service.TripService
}

func NewGRPChandler(server *grpc.Server, service *service.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, ptr *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {

	pickup := &types.Coordinate{
		Latitude:  ptr.StartLocation.Latitude,
		Longitude: ptr.StartLocation.Longitude,
	}

	destination := &types.Coordinate{
		Latitude:  ptr.EndLocation.Latitude,
		Longitude: ptr.EndLocation.Longitude,
	}

	t, err := h.service.GetRoute(ctx, pickup, destination)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route. err: %v", err)
	}

	estimatedFares, err := h.service.EstimatePackagesPriceWithRoute(t)
	if err != nil {
		log.Printf("[gRPCHandler.priviewTrip] Error estimating fares: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to estimate route fares. err: %v", err)
	}

	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, ptr.UserID)
	if err != nil {
		log.Printf("[gRPCHandler.priviewTrip] Error generating trip fares: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to generate route fares. err: %v", err)
	}

	faresProto := domain.ToRideFaresProto(fares)

	return &pb.PreviewTripResponse{
		Route:     t.ToProto(),
		RideFares: faresProto,
	}, nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, in *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	rideFareInCents, err := strconv.ParseFloat(in.RideFare, 64)
	if err != nil {
		log.Printf("[gRPC.CreateTrip] Error parsing rideFare: %v", err)
		return nil, err
	}

	rideFare := &domain.RideFareModel{
		ID:                primitive.NewObjectID(),
		UserId:            in.UserID,
		PackageSlug:       "van",
		TotalPriceInCents: rideFareInCents,
	}

	tripModel, err := h.service.CreateTrip(ctx, rideFare)
	if err != nil {
		log.Printf("[gRPC.CreateTrip] Error creating trip: %v", err)
		return nil, err
	}

	return &pb.CreateTripResponse{
		TripID: tripModel.ID.Hex(),
	}, nil
}

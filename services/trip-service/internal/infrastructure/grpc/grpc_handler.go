package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/service"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

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

	return &pb.PreviewTripResponse{
		Route:     t.ToProto(),
		RideFares: []*pb.RideFare{},
	}, nil
}

package grpc

import (
	"context"
	"ride-sharing/services/trip-service/internal/service"
	pb "ride-sharing/shared/proto/trip"

	"google.golang.org/grpc"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service service.TripService
}

func NewGRPChandler(server *grpc.Server, service service.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(context.Context, *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	return nil, nil
}

package main

import (
	pb_driver "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *previewTripRequest) toProto() *pb.PreviewTripRequest {
	return &pb.PreviewTripRequest{
		UserID: p.UserID,
		StartLocation: &pb.Coordinate{
			Latitude:  p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		EndLocation: &pb.Coordinate{
			Latitude:  p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type startTripRequest struct {
	RideFareID string `json:"rideFareID"`
	UserID     string `json:"userID"`
}

func (s *startTripRequest) toProto() *pb.CreateTripRequest {
	return &pb.CreateTripRequest{
		UserID:   s.UserID,
		RideFare: s.RideFareID,
	}
}

type startTripResponse struct {
	TripID string `json:"tripID"`
}

type driverRegisterRequest struct {
	ID          string `json:"id"`
	PackageSlug string `json:"packageSlug"`
}

func (d *driverRegisterRequest) toProto() *pb_driver.RegisterDriverRequest {
	return &pb_driver.RegisterDriverRequest{
		DriverID:    d.ID,
		PackageSlug: d.PackageSlug,
	}
}

type driverUnRegisterRequest struct {
	ID          string `json:"id"`
	PackageSlug string `json:"packageSlug"`
}

func (d *driverUnRegisterRequest) toProto() *pb_driver.RegisterDriverRequest {
	return &pb_driver.RegisterDriverRequest{
		DriverID:    d.ID,
		PackageSlug: d.PackageSlug,
	}
}

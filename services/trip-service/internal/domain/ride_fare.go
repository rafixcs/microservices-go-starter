package domain

import (
	pb "ride-sharing/shared/proto/trip"

	tripTypes "ride-sharing/services/trip-service/pkg/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserId            string
	PackageSlug       string // e.g.: van, luxury, sedan
	TotalPriceInCents float64
	Route             *tripTypes.OsrmApiResponse
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserId,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	faresProto := make([]*pb.RideFare, len(fares))
	for i, fare := range fares {
		faresProto[i] = fare.ToProto()
	}

	return faresProto
}

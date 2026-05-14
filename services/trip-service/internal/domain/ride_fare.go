package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FareRepository interface {
	CreateFare(ctx context.Context, fare *RideFareModel) (*RideFareModel, error)
}

type RideFareModel struct {
	ID                primitive.ObjectID
	UserId            string
	PackageSlug       string // e.g.: van, luxury, sedan
	TotalPriceInCents float64
}

package domain

type TripDriver struct {
	Id             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	ProfilePicture string `json:"profilePicture,omitempty"`
	CarPlate       string `json:"carPlate,omitempty"`
}

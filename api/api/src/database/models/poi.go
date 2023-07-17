package models

type PointOfInterest struct {
	BaseModel
	Type string  `json:"type" gorm:"uniqueIndex:unique_type_name"`
	Name string  `json:"name" gorm:"uniqueIndex:unique_type_name"`
	Lat  float64 `json:"lat"` // Lat is the latitude in decimal degrees.
	Lon  float64 `json:"lon"` // Lon is the longitude in decimal degrees.
}

func (PointOfInterest) TableName() string {
	return "points_of_interest" // otherwise it becomes `point_of_interests` 😬
}

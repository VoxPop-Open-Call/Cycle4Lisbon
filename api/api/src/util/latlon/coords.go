package latlon

import "math"

const (
	earthRadiusKm = 6371
)

type Coords struct {
	Lat float64 // Latitude in decimal degrees.
	Lon float64 // Longitude in decimal degrees.
}

// Dist calculates the distance in kilometers between two coordinates.
func Dist(c1, c2 Coords) float64 {
	dLat := degToRad(c1.Lat - c2.Lat)
	dLong := degToRad(c1.Lon - c2.Lon)

	lat1 := degToRad(c1.Lat)
	lat2 := degToRad(c2.Lat)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLong/2)*math.Sin(dLong/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

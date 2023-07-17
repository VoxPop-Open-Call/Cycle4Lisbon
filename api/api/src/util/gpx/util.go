package gpx

import (
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
)

// Speed calculates the average speed in km/h.
func Speed(dst float64, t time.Duration) float64 {
	if t.Hours() == 0 {
		return 0
	}
	return dst / t.Hours()
}

func distance(a, b Point) float64 {
	return latlon.Dist(latlon.Coords{
		Lat: a.Lat,
		Lon: a.Lon,
	}, latlon.Coords{
		Lat: b.Lat,
		Lon: b.Lon,
	})
}

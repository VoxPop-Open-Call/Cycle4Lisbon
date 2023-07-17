// Package gpx implements methods to parse, unmarshal and perform calculations
// on GPS Exchange Format data.
//
// https://en.wikipedia.org/wiki/GPS_Exchange_Format
//
// Source of the test files: https://github.com/gps-touring/sample-gpx
package gpx

import (
	"encoding/xml"
	"sort"
	"time"
)

// IdleSpeedThreshold is the minimum average speed (km/h) of an interval for it
// to be considered in motion (not idle).
var IdleSpeedThreshold = 1.0

type GPX struct {
	Metadata  Metadata `xml:"metadata"`
	WayPoints []Point  `xml:"wpt"`
	Track     Track    `xml:"trk"`
}

type Metadata struct {
	Name string `xml:"name"`
	Desc string `xml:"desc"`
}

type Track struct {
	Name    string  `xml:"name"`
	Desc    string  `xml:"desc"`
	Segment []Point `xml:"trkseg>trkpt"`
}

type Point struct {
	Name string     `xml:"name"`
	Desc string     `xml:"desc"`
	Lat  float64    `xml:"lat,attr"`
	Lon  float64    `xml:"lon,attr"`
	Ele  float64    `xml:"ele"`
	Time *time.Time `xml:"time"`
}

func (gpx *GPX) Unmarshal(data []byte) error {
	err := xml.Unmarshal(data, gpx)
	sort.SliceStable(gpx.Track.Segment, func(i, j int) bool {
		if gpx.Track.Segment[i].Time == nil || gpx.Track.Segment[j].Time == nil {
			return false
		}
		return gpx.Track.Segment[i].Time.Before(*gpx.Track.Segment[j].Time)
	})

	return err
}

// Distance calculates the total distance in kilometers of the track segment.
func (gpx *GPX) Distance() float64 {
	dist := 0.0
	for i := 0; i < len(gpx.Track.Segment)-1; i++ {
		dist += distance(
			gpx.Track.Segment[i],
			gpx.Track.Segment[i+1],
		)
	}
	return dist
}

// Duration calculates both the total and in motion time durations.
// An interval is considered idle if the speed is less than the
// IdleTimeThreshold.
func (gpx *GPX) Duration() (total, inMotion time.Duration) {
	pts := gpx.Track.Segment
	total, inMotion = 0, 0
	for i := 0; i < len(pts)-1; i++ {
		if pts[i].Time == nil || pts[i+1].Time == nil {
			continue
		}

		dst := distance(pts[i], pts[i+1])

		dur := pts[i+1].Time.Sub(*pts[i].Time)
		speed := dst / dur.Hours()

		total += dur
		if speed >= IdleSpeedThreshold {
			inMotion += dur
		}
	}

	return total, inMotion
}

// ElevationDelta is the difference in elevation between the ending and starting
// points.
func (gpx *GPX) ElevationDelta() float64 {
	return gpx.Track.Segment[len(gpx.Track.Segment)-1].Ele -
		gpx.Track.Segment[0].Ele
}

// MaxSpeed calculates the max speed between two contiguous points.
func (gpx *GPX) MaxSpeed() float64 {
	pts := gpx.Track.Segment
	max := 0.0
	for i := 0; i < len(pts)-1; i++ {
		if pts[i].Time == nil || pts[i+1].Time == nil {
			continue
		}

		dst := distance(pts[i], pts[i+1])

		dur := pts[i+1].Time.Sub(*pts[i].Time)
		speed := dst / dur.Hours()
		if speed > max {
			max = speed
		}
	}
	return max
}

func (gpx *GPX) StartPoint() *Point {
	if len(gpx.Track.Segment) == 0 {
		return nil
	}
	return &gpx.Track.Segment[0]
}

func (gpx *GPX) EndPoint() *Point {
	if len(gpx.Track.Segment) == 0 {
		return nil
	}
	return &gpx.Track.Segment[len(gpx.Track.Segment)-1]
}

package gpx

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Source of the GPX test data: https://github.com/gps-touring/sample-gpx

func TestGPX(t *testing.T) {
	var timeMustParse = func(v string) *time.Time {
		tm, err := time.Parse(time.RFC3339, v)
		require.NoError(t, err)
		return &tm
	}

	testCases := []struct {
		file       string
		dist       float64
		total      time.Duration
		inMotion   time.Duration
		aveSpeed   float64
		eleDelta   float64
		maxSpeed   float64
		startPoint *Point
		endPoint   *Point
	}{
		{
			file:     "./testdata/Southampton_Portsmouth.gpx",
			dist:     39.5,
			eleDelta: -12.40,
			startPoint: &Point{
				Lat: 50.90971,
				Lon: -1.40435,
				Ele: 19.956,
			},
			endPoint: &Point{
				Lat: 50.8117,
				Lon: -1.0862,
				Ele: 7.56,
			},
		},
		{
			file:     "./testdata/Lannion_Plestin_parcours24.gpx",
			dist:     23.74,
			total:    2*time.Hour + 8*time.Minute + 25*time.Second,
			inMotion: 1*time.Hour + 36*time.Minute + 9*time.Second,
			aveSpeed: 14.8,
			eleDelta: -6.8,
			maxSpeed: 119.8,
			startPoint: &Point{
				Lat:  48.73168783262372,
				Lon:  -3.462917329743505,
				Ele:  9.300000000000001,
				Time: timeMustParse("2013-03-08T11:06:01Z"),
			},
			endPoint: &Point{
				Lat:  48.670295337215066,
				Lon:  -3.637969940900803,
				Ele:  2.5,
				Time: timeMustParse("2013-03-08T13:14:26Z"),
			},
		},
		{
			file:     "./testdata/Trebeurden_Lannion_parcours13.gpx",
			dist:     13.4,
			total:    2*time.Hour + 0*time.Minute + 55*time.Second,
			inMotion: 0*time.Hour + 58*time.Minute + 0*time.Second,
			aveSpeed: 13.8,
			eleDelta: -66.3,
			maxSpeed: 37.2,
			startPoint: &Point{
				Lat:  48.766598459333181,
				Lon:  -3.562515210360289,
				Ele:  75.599999999999994,
				Time: timeMustParse("2013-03-08T09:05:06Z"),
			},
			endPoint: &Point{
				Lat:  48.73168783262372,
				Lon:  -3.462917329743505,
				Ele:  9.300000000000001,
				Time: timeMustParse("2013-03-08T11:06:01Z"),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.file, func(t *testing.T) {
			data, err := os.ReadFile(tC.file)
			require.NoError(t, err)

			gpx := new(GPX)
			require.NoError(t, gpx.Unmarshal(data))

			dist := gpx.Distance()
			assert.Truef(t, math.Abs(tC.dist-dist) < 0.1,
				"distance '%f' not in margin of error of %f", dist, tC.dist)

			total, inMotion := gpx.Duration()
			assert.Truef(t, math.Abs(float64(tC.total-total)) < 0.1,
				"duration '%v' not in margin of error of %v", total, tC.total)
			assert.Truef(t, math.Abs(float64(tC.inMotion-inMotion)) < 0.1,
				"duration '%v' not in margin of error of %v", inMotion, tC.inMotion)

			aveSpeedInMotion := Speed(dist, inMotion)
			assert.Truef(t, math.Abs(tC.aveSpeed-aveSpeedInMotion) < 0.1,
				"average speed '%f' not in margin of error of %f",
				aveSpeedInMotion, tC.aveSpeed)

			maxSpeed := gpx.MaxSpeed()
			assert.Truef(t, math.Abs(tC.maxSpeed-maxSpeed) < 0.1,
				"maximum speed '%f' not in margin of error of %f",
				maxSpeed, tC.maxSpeed)

			eleDelta := gpx.ElevationDelta()
			assert.Truef(t, math.Abs(tC.eleDelta-eleDelta) < 0.1,
				"elevation delta '%f' not in margin of error of %f",
				eleDelta, tC.eleDelta)

			assert.Equal(t, tC.startPoint, gpx.StartPoint())
			assert.Equal(t, tC.endPoint, gpx.EndPoint())
		})
	}
}

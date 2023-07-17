package latlon

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Geocoder struct {
	token string
}

const (
	gcMapsReverseURLStr = "https://maps.googleapis.com/maps/api/geocode/json?" +
		"latlng=%f,%f&" +
		"key=%s&" +
		"result_type=street_address"

	gcOSMReverseURLStr = "https://nominatim.openstreetmap.org/reverse?" +
		"lat=%f&" +
		"lon=%f&" +
		"format=jsonv2&" +
		"addressdetails=1&" +
		"namedetails=1"
)

func NewGeocoder(token string) *Geocoder {
	return &Geocoder{token}
}

type reverseMapsResult struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			LocationType string `json:"location_type"`
		} `json:"geometry"`
		Types []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

func (rmr reverseMapsResult) simplifiedAddr() string {
	return rmr.Results[0].FormattedAddress
}

func (g *Geocoder) reverseMapsURL(coords Coords) string {
	return fmt.Sprintf(gcMapsReverseURLStr, coords.Lat, coords.Lon, g.token)
}

// reverseMaps requests the Google Maps API for reverse geocoding results.
func (g *Geocoder) reverseMaps(coords Coords) (*reverseMapsResult, error) {
	req, err := http.NewRequest(http.MethodGet, g.reverseMapsURL(coords), nil)
	if err != nil {
		return nil,
			fmt.Errorf("failed to create request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		data, err := io.ReadAll(res.Body)
		return nil,
			fmt.Errorf("response status %d: %s %v",
				res.StatusCode, string(data), err)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil,
			fmt.Errorf("failed to read response body: %v", err)
	}

	var rev reverseMapsResult
	if err = json.Unmarshal(raw, &rev); err != nil {
		return nil,
			fmt.Errorf("failed to parse response: %v", err)
	}

	return &rev, nil
}

type reverseOSMResult struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Address     struct {
		HouseNumber string `json:"house_number"`
		Square      string `json:"square"`
		Road        string `json:"road"`
		City        string `json:"city"`
		Town        string `json:"town"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	} `json:"address"`
	NameDetails struct {
		Name string `json:"name"`
	} `json:"namedetails"`
}

func (ror *reverseOSMResult) simplifiedAddr() string {
	sb := strings.Builder{}
	if ror.Name != "" {
		sb.WriteString(ror.Name)
	} else if ror.Address.HouseNumber != "" {
		sb.WriteString(ror.Address.HouseNumber)

		if ror.Address.Road != "" {
			sb.WriteString(" ")
			sb.WriteString(ror.Address.Road)
		}
	}

	if ror.Address.City != "" {
		sb.WriteString(", ")
		sb.WriteString(ror.Address.City)
	} else if ror.Address.Town != "" {
		sb.WriteString(", ")
		sb.WriteString(ror.Address.Town)
	}

	return sb.String()
}

func (g *Geocoder) reverseOSMURL(coords Coords) string {
	return fmt.Sprintf(gcOSMReverseURLStr, coords.Lat, coords.Lon)
}

func (g *Geocoder) reverseOSM(coords Coords) (*reverseOSMResult, error) {
	req, err := http.NewRequest(http.MethodGet, g.reverseOSMURL(coords), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		data, err := io.ReadAll(res.Body)
		return nil, fmt.Errorf("response status %d: %s %v",
			res.StatusCode, string(data), err)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil,
			fmt.Errorf("failed to read response body: %v", err)
	}

	var rev reverseOSMResult
	if err = json.Unmarshal(raw, &rev); err != nil {
		return nil,
			fmt.Errorf("failed to parse response data: %v", err)
	}

	return &rev, nil
}

// ReverseAddr does reverse geocoding of the given coordinates, returning a
// simplified address.
func (g *Geocoder) ReverseAddr(coords Coords) string {
	mapsRes, err := g.reverseMaps(coords)
	if err == nil && mapsRes.Status == "OK" {
		return mapsRes.simplifiedAddr()
	}

	log.Printf("failed to query Maps, falling back to OSM: %v, %s\n",
		err, mapsRes.Status)

	osmRes, err := g.reverseOSM(coords)
	if err == nil {
		return osmRes.simplifiedAddr()
	}

	log.Printf("failed to query OSM, no fallback: %v", err)

	return ""
}

package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	// eventsUrl is the URL to fetch events from.
	//
	// The API takes a bit less than a minute to return 100 results, which is
	// too close to the 60s timeout. This should be taken into consideration
	// when choosing the `per_page` record limit.
	//
	// There are some undocumented endpoints apart from this one that might be
	// useful in the future:
	// `/venues`: entire list of venues and metadata, including lat and long.
	// `/articles`
	// `/tax`: "taxonomy". Basically key-value pairs.
	//		`?tax=category`: corresponds the Subject field of the Event.
	//		`?tax=post_tag`: tags with a usage count.
	//		`?tax=event-venue`: useless, just call `/venues`.
	//		`?tax=accessibility`: not much here, only 3 entries...
	//		`?tax=target_audience`: not a lot here either.
	//		`?tax=free_tag`: the same as `post_tag`, some different tags.
	eventsUrl = "https://www.agendalx.pt/wp-json/agendalx/v1/events?per_page=50"
)

// Event is the type of the elements in the response to the eventsUrl.
//
// All strings are UTF-8 encoded.
//
// Since the API is a mess - some fields can have different types depending on
// the data - some values have to implement the json.Unmarshal interface, to
// allow parsing into a type we can work with.
type Event struct {
	ID uint64 `json:"id"`

	Title    title    `json:"title"`
	Subtitle subtitle `json:"subtitle"`
	// Subject of the event, like "visitas guiadas" or "artes". Can also have
	// multiple subjects separated by a comma: "ciência, visitas  guiadas".
	Subject string `json:"subject"`
	// FeaturedMedia is a url to an image. An empty string is returned in case
	// of no data.
	FeaturedMedia featuredMedia `json:"featured_media_large"`

	// Description of the event, truncated to around 200 bytes (less than 200
	// characters).
	//
	// The string may contain html tags which, because of the blind truncation,
	// may not be correctly closed or even complete. This value is probably
	// useless.
	Description description `json:"description"`

	// StringDates are descriptions of the dates of the event, in the format
	// "12 dezembro 2022 a 31 dezembro 2023".
	//
	// All the dates I encountered are in this format, but it's probably not a
	// good idea to count on it to parse them.
	StringDates []string `json:"string_dates"`
	// StringTimes is a free-form description of the times at which the event
	// occurs, for example: "sex: 21h30; sáb: 19h" or "vários horários".
	StringTimes string `json:"string_times"`

	Venue          maptype `json:"venue"`
	Categories     maptype `json:"categories_name_list"`
	Tags           maptype `json:"tags_name_list"`
	TargetAudience maptype `json:"target_audience"`
	Accessibility  maptype `json:"accessibility"`

	// Link is the URL of the full article.
	Link string `json:"link"`

	// Occurrences is a list of dates in the format "2022-12-12".
	Occurrences []string `json:"occurences"`
}

type title struct {
	// Rendered is a UTF encoded string of the title of the event.
	Rendered string `json:"rendered"`
}

// unmarshalStringOrArray parses data that can be an empty string or an array
// of strings.
//
// For some reason (and only for some of the fields), the API returns an empty
// string if there is no data, and an array of strings (with only one element,
// in most cases) if there is. This function returns nil in case of the former,
// and a slice in case of the latter.
func unmarshalStringOrArray(data []byte) ([]string, error) {
	if string(data) == "\"\"" {
		return nil, nil
	}
	var res []string
	err := json.Unmarshal(data, &res)
	return res, err
}

type subtitle []string

func (s *subtitle) UnmarshalJSON(data []byte) error {
	val, err := unmarshalStringOrArray(data)
	*s = val
	return err
}

type description []string

func (d *description) UnmarshalJSON(data []byte) error {
	val, err := unmarshalStringOrArray(data)
	*d = val
	return err
}

type featuredMedia string

func (f *featuredMedia) UnmarshalJSON(data []byte) error {
	if string(data) == "false" {
		*f = ""
		return nil
	}

	var s string
	err := json.Unmarshal(data, &s)
	*f = featuredMedia(s)
	return err
}

type maptype map[string]mapdata

type mapdata struct {
	ID   uint64 `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (v *maptype) UnmarshalJSON(data []byte) error {
	// An empty value is sometimes represented with a `false`, sometimes an
	// empty array, sometimes an empty string...
	// Who made this?!
	if string(data) == "false" ||
		string(data) == "[]" ||
		string(data) == "\"\"" {
		*v = make(maptype)
		return nil
	}

	res := make(map[string]mapdata)
	err := json.Unmarshal(data, &res)
	*v = res
	return err
}

// FetchEvents requests the events from the `agendalx.pt` API and returns them.
func FetchEvents() ([]Event, error) {
	req, err := http.NewRequest(http.MethodGet, eventsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("content-type", "application/json")

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

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var events []Event
	err = json.Unmarshal(data, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data: %v", err)
	}

	return events, nil
}

package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalEvents(t *testing.T) {
	for i, tc := range []struct {
		f   string // filename
		len int    // expected number of events
	}{
		{"./testdata/agendalx-2023_03_17_page1.json", 100},
		{"./testdata/agendalx-2023_03_17_page2.json", 100},
		{"./testdata/agendalx-2023_03_17_page3.json", 100},
		{"./testdata/agendalx-2023_03_17_page4.json", 100},
		{"./testdata/agendalx-2023_03_17_page5.json", 61},
		{"./testdata/agendalx-2023_04_13_page1.json", 90},
	} {
		data, err := os.ReadFile(tc.f)
		require.NoError(t, err)

		var events []Event
		err = json.Unmarshal(data, &events)
		require.NoError(t, err, "failed on test: %d", i)
		assert.Len(t, events, tc.len, "test case %d, got length: %d", i, len(events))
		for j, e := range events {
			assert.NotEmpty(t, e, "failed on test: %d %d", i, j)
			assert.NotEmpty(t, e.Description)
			assert.NotEmpty(t, e.Title)
			assert.NotEmpty(t, e.Title.Rendered)
		}
	}
}

func TestMapTypeJSON(t *testing.T) {
	data := `{
      "museu-nacional-da-musica": {
        "id": 824,
        "slug": "museu-nacional-da-musica",
        "name": "Museu Nacional da M\u00fasica"
      }
    }`

	var venue maptype
	err := venue.UnmarshalJSON([]byte(data))

	assert.NoError(t, err)
	assert.NotEmpty(t, venue)
	key := "museu-nacional-da-musica"
	assert.Contains(t, venue, key)
	assert.Equal(t, uint64(824), venue[key].ID)
	assert.Equal(t, key, venue[key].Slug)
	assert.Equal(t, "Museu Nacional da M\u00fasica", venue[key].Name)
}

func TestFetchEvents(t *testing.T) {
	data, err := os.ReadFile("./testdata/agendalx-2023_03_17_page1.json")
	require.NoError(t, err)
	var expected []Event
	err = json.Unmarshal(data, &expected)
	require.NoError(t, err)

	// Setup a mock server.
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/mock/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
	srv := &http.Server{
		Addr:    ":8080",
		Handler: srvMux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()
	defer srv.Close()

	// Override the URL, to fetch events from the mock server instead.
	eventsUrl = "http://localhost:8080/mock/events"

	// Retry fetching the events until the server is available.
	var events []Event
	for i := 0; i < 10; i++ {
		events, err = FetchEvents()
		if err != nil {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	require.NoError(t, err)
	require.NotEmpty(t, events)

	assert.Len(t, events, 100)
	for i, e := range events {
		assert.NotEmpty(t, e, "failed on test %d", i)
	}

	assert.Equal(t, expected, events)
}

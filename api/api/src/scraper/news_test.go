package scraper

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestParseNews(t *testing.T) {
	for i, tc := range []struct {
		f   string // filename
		len int    // expected number of news entries
	}{
		{"./testdata/lisboa_news-2023-04-19_page1.html", 20},
		{"./testdata/lisboa_news-2023-04-19_page2.html", 20},
		{"./testdata/lisboa_news-2023-04-19_page3.html", 20},
		{"./testdata/lisboa_news-2023-04-19_page4.html", 20},
	} {
		file, err := os.Open(tc.f)
		require.NoError(t, err, "failed on test %d", i)

		news, err := parseNews(file)
		require.NoError(t, err, "failed on test %d", i)
		require.NotEmpty(t, news, "failed on test %d", i)

		assert.Len(t, news, tc.len, "failed on test %d", i)

		for _, item := range news {
			assert.NotEmpty(t, item.Title)
			assert.NotEmpty(t, item.Image)
			assert.NotEmpty(t, item.Tags)
			assert.NotEmpty(t, item.Date)
			assert.NotEmpty(t, item.Link)
		}
	}
}

func TestParseNewsItem(t *testing.T) {
	tokens := html.NewTokenizer(strings.NewReader(`
	<div class="card">
		<div class="card-img-top">
			<a
				href="https://www.lisboa.pt/atualidade/noticias/detalhe/animais-em-risco-vao-ter-cuidados-medico-veterinarios-gratuitos/"
				><img
					src="https://www.lisboa.pt/fileadmin/atualidade/noticias/user_upload/Armindo-3270.jpg"
			/></a>
		</div>
		<div class="card-body d-flex flex-column">
			<div class="d-flex justify-content-between mb-3">
				<div class="cat">
					<span>Bem-Estar Animal</span><span>Direitos Sociais</span>
				</div>
				<time class="small">18.04.2023</time>
			</div>
			<h3>
				<a
					href="https://www.lisboa.pt/atualidade/noticias/detalhe/animais-em-risco-vao-ter-cuidados-medico-veterinarios-gratuitos/"
					>Animais em risco vão ter cuidados médico-veterinários
					gratuitos</a
				>
			</h3>
		</div>
	</div>
	`))

	newsItem, err := parseNewsItem(tokens)
	require.NoError(t, err)
	assert.Equal(t, News{
		Title: "Animais em risco vão ter cuidados médico-veterinários gratuitos",
		Image: "https://www.lisboa.pt/fileadmin/atualidade/noticias/user_upload/Armindo-3270.jpg",
		Tags:  []string{"Bem-Estar Animal", "Direitos Sociais"},
		Date:  "18.04.2023",
		Link:  "https://www.lisboa.pt/atualidade/noticias/detalhe/animais-em-risco-vao-ter-cuidados-medico-veterinarios-gratuitos/",
	}, newsItem)

}

func TestFetchNews(t *testing.T) {
	data, err := os.ReadFile("./testdata/lisboa_news-2023-04-19_page1.html")
	require.NoError(t, err)

	file, err := os.Open("./testdata/lisboa_news-2023-04-19_page1.html")
	require.NoError(t, err)
	expected, err := parseNews(file)
	require.NoError(t, err)

	// Setup a mock server.
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/mock/news", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
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

	// Override the URL, to fetch news from the mock server instead.
	newsUrl = "http://localhost:8080/mock/news"

	// Retry fetching the news until the server is available.
	var news []News
	for i := 0; i < 10; i++ {
		news, err = FetchNews()
		if err != nil {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	require.NoError(t, err)
	require.NotEmpty(t, news)

	assert.Len(t, news, 20)
	for i, e := range news {
		assert.NotEmpty(t, e, "failed on test %d", i)
	}

	assert.Equal(t, expected, news)
}

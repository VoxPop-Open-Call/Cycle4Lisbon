package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

var (
	newsUrl = "https://www.lisboa.pt/typo3conf/ext/dmc_tables_extend/Classes/Noticias/noticias.php"
	// pageItemCount is how many news items are in a page, by default.
	pageItemCount = 20
)

type News struct {
	Title string
	Image string
	Tags  []string
	Date  string
	Link  string
}

// parseNewsItem creates a News struct by parsing data from the tokenizer,
// assuming the current tag is the start of a news item.
func parseNewsItem(tokens *html.Tokenizer) (News, error) {
	isImgToken := func(token html.Token) bool {
		return token.Data == "img" && len(token.Attr) == 1 &&
			token.Attr[0].Key == "src"
	}

	isLinkToken := func(token html.Token) bool {
		return token.Data == "a" && len(token.Attr) == 1 &&
			token.Attr[0].Key == "href"
	}

	isCatSpan := func(token html.Token) bool {
		return token.Data == "span" && len(token.Attr) == 0
	}

	isTimeTag := func(token html.Token) bool {
		return token.Data == "time"
	}

	isH3Tag := func(token html.Token) bool {
		return token.Data == "h3"
	}

	cleanupTitle := func(title string) string {
		return strings.Join(strings.Fields(title), " ")
	}

	result := News{}

	inCat := false
	inTime := false
	inH3 := false
	inTitle := false

loop:
	for {
		tt := tokens.Next()
		switch tt {
		case html.ErrorToken:
			return News{}, errors.New("unexpected end-of-file")
		case html.StartTagToken:
			t := tokens.Token()
			if isImgToken(t) {
				result.Image = t.Attr[0].Val
			} else if isLinkToken(t) {
				result.Link = t.Attr[0].Val
				inTitle = inH3
			} else if isCatSpan(t) {
				inCat = true
			} else if isTimeTag(t) {
				inTime = true
			} else if isH3Tag(t) {
				inH3 = true
			}
		case html.SelfClosingTagToken:
			t := tokens.Token()
			if isImgToken(t) {
				result.Image = t.Attr[0].Val
			}
		case html.TextToken:
			if inCat {
				result.Tags = append(result.Tags, tokens.Token().Data)
				inCat = false
			} else if inTime {
				result.Date = tokens.Token().Data
				inTime = false
			} else if inTitle {
				result.Title = cleanupTitle(tokens.Token().Data)
				inTitle = false
				inH3 = false

				break loop // stop searching after the Title
			}
		}
	}

	return result, nil
}

// parseNews parses the raw html into a slice of News.
func parseNews(r io.Reader) ([]News, error) {
	// isNewsParentToken returns whether the token is the parent tag of each
	// news item.
	isNewsParentToken := func(token html.Token) bool {
		return token.Data == "div" && len(token.Attr) == 1 &&
			token.Attr[0].Key == "class" &&
			token.Attr[0].Val == "card-group-element-item event-image"
	}

	news := make([]News, 0, pageItemCount)
	htmlTokens := html.NewTokenizer(r)
loop:
	for {
		tt := htmlTokens.Next()
		switch tt {
		case html.ErrorToken:
			// No more tokens
			break loop
		case html.StartTagToken:
			t := htmlTokens.Token()
			if isNewsParentToken(t) {
				n, err := parseNewsItem(htmlTokens)
				if err != nil {
					return nil, err
				}
				news = append(news, n)
			}
		}
	}

	return news, nil
}

// FetchNews requests the news from the `lisboa.pt` website, parses and returns
// them.
func FetchNews() ([]News, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		newsUrl,
		bytes.NewReader([]byte("inicio=inicio&lingua=0")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

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

	return parseNews(res.Body)
}

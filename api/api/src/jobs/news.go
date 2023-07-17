package jobs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/types"
	"bitbucket.org/pensarmais/cycleforlisbon/src/scraper"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/getsentry/sentry-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// News job names.
const (
	// Retrieve news using the scraper and store them in the database.
	FetchNews = "fetch-news"
)

// Time between fetching news.
//
// There are usualy less than 3 news per day, so fetching once a day should
// ensure none are missed.
const fetchNewsPeriod = 24 * time.Hour

func fetchNews(wrkr *worker.Worker, db *gorm.DB) *worker.Job {
	reschedule := func() {
		if err := wrkr.Schedule(&worker.TaskConfig{
			JobName:     FetchNews,
			ScheduledTo: time.Now().Add(fetchNewsPeriod),
		}); err != nil {
			log.Printf("failed to reschedule news fetching: %v", err)
		}
	}

	parseDate := func(src string) (*types.Date, error) {
		t, err := time.Parse("02.01.2006", src)
		if err != nil {
			return nil, err
		}
		date := new(types.Date)
		err = date.Scan(t)
		return date, err
	}

	toExtContentModel := func(news scraper.News) (models.ExternalContent, error) {
		date, err := parseDate(news.Date)
		return models.ExternalContent{
			Type:         "news",
			Title:        news.Title,
			ImageUrl:     news.Image,
			ArticleUrl:   news.Link,
			Subject:      strings.ToLower(strings.Join(news.Tags, ", ")),
			Date:         date,
			LanguageCode: "pt",
		}, err
	}

	return &worker.Job{
		Name: FetchNews,
		Handler: func(ctx context.Context, _ []byte) error {
			news, err := scraper.FetchNews()
			if err != nil {
				err = fmt.Errorf("failed to fetch news: %v", err)
				sentry.CaptureException(err)
				return err
			}

			dbNews := make([]models.ExternalContent, len(news))
			for i, n := range news {
				dbNews[i], err = toExtContentModel(n)
				if err != nil {
					err = fmt.Errorf("error converting to news model: %v", err)
					sentry.CaptureException(err)
					return err
				}
			}

			return db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "article_url"}},
				DoNothing: true,
			}).Create(&dbNews).Error
		},
		OnSuccess: reschedule,
		OnFailure: reschedule,
	}
}

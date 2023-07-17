package jobs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/scraper"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/getsentry/sentry-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Event job names.
const (
	// Retrieve events using the scraper and store them in the database.
	FetchEvents = "fetch-events"
)

// Time between fetching events.
const fetchEventsPeriod = 24 * time.Hour

func fetchEvents(wrkr *worker.Worker, db *gorm.DB) *worker.Job {
	reschedule := func() {
		if err := wrkr.Schedule(&worker.TaskConfig{
			JobName:     FetchEvents,
			ScheduledTo: time.Now().Add(fetchEventsPeriod),
		}); err != nil {
			log.Printf("failed to reschedule event fetching: %v", err)
		}
	}

	toExtContentModel := func(event scraper.Event) models.ExternalContent {
		return models.ExternalContent{
			Type:         "event",
			Title:        event.Title.Rendered,
			Subtitle:     strings.Join(event.Subtitle, "\n"),
			ImageUrl:     string(event.FeaturedMedia),
			ArticleUrl:   event.Link,
			Subject:      strings.ToLower(event.Subject),
			Description:  strings.Join(event.Description, "\n"),
			Period:       strings.Join(event.StringDates, ", "),
			Time:         event.StringTimes,
			LanguageCode: "pt",
		}
	}

	return &worker.Job{
		Name: FetchEvents,
		Handler: func(ctx context.Context, _ []byte) error {
			events, err := scraper.FetchEvents()
			if err != nil {
				err = fmt.Errorf("failed to fetch events: %v", err)
				sentry.CaptureException(err)
				return err
			}

			dbEvents := make([]models.ExternalContent, len(events))
			for i, e := range events {
				dbEvents[i] = toExtContentModel(e)
			}

			return db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "article_url"}},
				DoNothing: true,
			}).Create(&dbEvents).Error
		},
		OnSuccess: reschedule,
		OnFailure: reschedule,
	}
}

package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"gorm.io/gorm"
)

// Password related job names.
const (
	// Delete old password reset codes from the database.
	PasswordResetCodeCleanup = "pwd-rst-code-cleanup"
)

const (
	// How long to keep password reset codes in the database for, after they
	// expire.
	passwordResetCodeLifetime = 7 * 24 * time.Hour
)

// Time between cleanup tasks.
const passwordResetCodeCleanupPeriod = 24 * time.Hour

func passwordResetCodeCleanup(wrkr *worker.Worker, db *gorm.DB) *worker.Job {
	reschedule := func() {
		if err := wrkr.Schedule(&worker.TaskConfig{
			JobName:     PasswordResetCodeCleanup,
			ScheduledTo: time.Now().Add(passwordResetCodeCleanupPeriod),
		}); err != nil {
			log.Printf("failed to reschedule password reset code cleanup: %v",
				err)
		}
	}

	return &worker.Job{
		Name: PasswordResetCodeCleanup,
		Handler: func(ctx context.Context, _ []byte) error {
			deleted, err := query.PasswordResetCodes.
				DeleteOlderThan(passwordResetCodeLifetime, db)

			if err != nil {
				return fmt.Errorf(
					"failed to delete old password reset codes: %v", err,
				)
			}

			log.Printf("%s: deleted %d password reset codes",
				PasswordResetCodeCleanup, deleted)
			return nil
		},
		OnSuccess: reschedule,
		OnFailure: reschedule,
	}

}

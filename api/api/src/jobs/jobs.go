// Package jobs defines Jobs to be run by the job queue.
// See package `worker`.
package jobs

import (
	"bitbucket.org/pensarmais/cycleforlisbon/src/achievements"
	"bitbucket.org/pensarmais/cycleforlisbon/src/firebase"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"gorm.io/gorm"
)

// All initializes and returns all jobs, to be registered with a worker.
func All(
	wrkr *worker.Worker,
	fbase *firebase.Client,
	db *gorm.DB,
	achs *achievements.Service,
	host string,
) []*worker.Job {
	return []*worker.Job{
		fetchNews(wrkr, db),
		fetchEvents(wrkr, db),
		fcmNotify(fbase.Fcm, db),
		fcmMulticast(fbase.Fcm, db),
		fcmCleanup(wrkr, fbase.Fcm, db),
		passwordResetCodeCleanup(wrkr, db),
		updateAchievements(achs, fbase.Fcm, wrkr, db, host),
	}
}

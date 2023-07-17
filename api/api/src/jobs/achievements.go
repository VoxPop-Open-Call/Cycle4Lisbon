package jobs

import (
	"context"
	"fmt"

	"bitbucket.org/pensarmais/cycleforlisbon/src/achievements"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/firebase"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gobutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// Update the achievements of a user.
	// Args are of type `UpdateAchievementsArgs`.
	UpdateAchievements = "achievements-update"
)

type UpdateAchievementsArgs struct {
	UserID uuid.UUID
	State  achievements.State
}

func updateAchievements(
	achs *achievements.Service,
	fbase fcmSender,
	wrkr *worker.Worker,
	db *gorm.DB,
	host string,
) *worker.Job {
	argsCodec := gobutil.NewGobCodec[UpdateAchievementsArgs]()
	msgCodec := gobutil.NewGobCodec[messaging.Message]()

	return &worker.Job{
		Name: UpdateAchievements,
		Handler: func(ctx context.Context, raw []byte) error {
			if args, err := argsCodec.Decode(raw); err != nil {
				return fmt.Errorf("failed to decode args: %v", err)

			} else if newAchs, err := achs.
				Update(args.UserID, args.State); err != nil {
				return fmt.Errorf("failed to update user achievements: %v", err)

			} else if tokens, err := query.FCMTokens.
				Of(args.UserID.String(), db); err != nil {
				return fmt.Errorf("failed to retrieve user fcm tokens: %v", err)

			} else {
				for _, ach := range newAchs {
					for _, token := range tokens {
						raw, err := msgCodec.Encode(firebase.NewAchievementMessage(
							token, ach.Achievement, host,
						))
						if err != nil {
							return fmt.Errorf("failed to encode args: %v", err)
						}

						wrkr.Schedule(&worker.TaskConfig{
							JobName: FcmNotify,
							Args:    raw,
						})
					}
				}
			}

			return nil
		},
	}
}

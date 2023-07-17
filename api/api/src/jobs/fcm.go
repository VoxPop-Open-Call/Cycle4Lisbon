package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/firebase"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/gobutil"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"firebase.google.com/go/v4/messaging"
	"gorm.io/gorm"
)

// FCM job names.
const (
	// Notify a single token or topic.
	// Args are of type `Message`.
	FcmNotify = "fcm-notify"

	// Notify multiple tokens.
	// Args are of type `MulticastMessage`.
	FcmMulticast = "fcm-multicast"

	// Delete old and stale tokens.
	FcmTokenCleanup = "fcm-token-cleanup"
)

// Thresholds for detecting and deleting stale tokens.
const (
	tokenFailureLimit    = 20
	tokenInactivityLimit = time.Hour * 24 * 60
)

// Time between token cleanup tasks.
const tokenCleanupPeriod = 24 * time.Hour

type fcmSender interface {
	Send(context.Context, *messaging.Message) (string, error)
}

type fcmMulticaster interface {
	SendMulticast(
		context.Context,
		*messaging.MulticastMessage,
	) (*messaging.BatchResponse, error)
}

type fcmClient interface {
	fcmSender
	fcmMulticaster

	SendMulticastDryRun(
		context.Context,
		*messaging.MulticastMessage,
	) (*messaging.BatchResponse, error)
}

func fcmNotify(fcm fcmSender, db *gorm.DB) *worker.Job {
	ed := gobutil.NewGobCodec[messaging.Message]()

	return &worker.Job{
		Name: FcmNotify,
		Handler: func(ctx context.Context, raw []byte) error {
			msg, err := ed.Decode(raw)
			if err != nil {
				return fmt.Errorf("failed to decode args: %v", err)
			}

			res, err := fcm.Send(ctx, &msg)
			if err != nil {
				if len(msg.Token) > 0 {
					incrementFailureCount(msg.Token, db)
				}
				return fmt.Errorf("error sending message: %v", err)
			}

			if len(msg.Token) > 0 {
				resetFailureCount(msg.Token, db)
			}

			log.Println(FcmNotify, "successfully sent message:", res)
			return nil
		},
		Retries:  6,
		Delay:    time.Minute,
		MaxDelay: 5 * time.Hour, // about 7 hours between the first and last try
	}
}

func fcmMulticast(fcm fcmMulticaster, db *gorm.DB) *worker.Job {
	ed := gobutil.NewGobCodec[messaging.MulticastMessage]()

	return &worker.Job{
		Name: FcmMulticast,
		Handler: func(ctx context.Context, raw []byte) error {
			msg, err := ed.Decode(raw)
			if err != nil {
				return fmt.Errorf("failed to decode args: %v", err)
			}

			br, err := fcm.SendMulticast(ctx, &msg)
			if err != nil {
				return fmt.Errorf("error sending message: %v", err)
			}

			for i, res := range br.Responses {
				token := msg.Tokens[i]
				if res.Success {
					resetFailureCount(token, db)
				} else {
					incrementFailureCount(token, db)
					// TODO: schedule a Notify task for each failed token?
				}
			}

			log.Printf(
				"%s: successfully sent %d messages, failed %d",
				FcmMulticast,
				br.SuccessCount,
				br.FailureCount,
			)
			return nil
		},
		Retries:  6,
		Delay:    time.Minute,
		MaxDelay: 5 * time.Hour, // about 7 hours between the first and last try
	}
}

func fcmCleanup(wrkr *worker.Worker, fcm fcmClient, db *gorm.DB) *worker.Job {
	reschedule := func() {
		if err := wrkr.Schedule(&worker.TaskConfig{
			JobName:     FcmTokenCleanup,
			ScheduledTo: time.Now().Add(tokenCleanupPeriod),
		}); err != nil {
			log.Printf("failed to reschedule token cleanup: %v", err)
		}
	}

	tokenStrings := func(dbTokens []models.FCMToken) []string {
		result := make([]string, len(dbTokens))
		for i, t := range dbTokens {
			result[i] = t.Token
		}
		return result
	}

	// Validate the tokens, delete the ones that fail.
	// Returns the number of deleted tokens.
	validate := func(ctx context.Context, tokens []models.FCMToken) (int, error) {
		deleted := 0
		msg := firebase.NewTestMessage(tokenStrings(tokens))
		br, err := fcm.SendMulticastDryRun(ctx, &msg)
		if err != nil {
			return 0, fmt.Errorf("error sending test message: %v", err)
		}

		for i, res := range br.Responses {
			token := tokens[i].Token
			if res.Success {
				resetFailureCount(token, db)
			} else {
				delete(token, db)
				deleted++
			}
		}
		return deleted, nil
	}

	return &worker.Job{
		Name: FcmTokenCleanup,
		Handler: func(ctx context.Context, _ []byte) error {
			deleted, err := query.FCMTokens.
				DeleteAllInactive(tokenInactivityLimit, db)
			if err != nil {
				return fmt.Errorf("failed to delete inactive tokens: %v", err)
			}

			suspectTokens, err := query.FCMTokens.GetFailed(tokenFailureLimit, db)
			if err != nil {
				return fmt.Errorf("error querying failed tokens: %v", err)
			}

			if len(suspectTokens) > 0 {
				susDeleted, err := validate(ctx, suspectTokens)
				if err != nil {
					return err
				}
				deleted += susDeleted
			}

			log.Printf("%s: deleted %d tokens", FcmTokenCleanup, deleted)
			return nil
		},
		OnSuccess: reschedule,
		OnFailure: reschedule,
	}
}

func incrementFailureCount(token string, db *gorm.DB) {
	err := query.FCMTokens.IncrementFailureCount(token, db)

	if err != nil {
		log.Printf("error incrementing token failure count: %v", err)
	}
}

func resetFailureCount(token string, db *gorm.DB) {
	err := query.FCMTokens.ResetFailureCount(token, db)

	if err != nil {
		log.Printf("error resetting token failure count: %v", err)
	}
}

func delete(token string, db *gorm.DB) {
	err := query.FCMTokens.Delete(token, db)

	if err != nil {
		log.Printf("error deleting token: %v", err)
	}
}

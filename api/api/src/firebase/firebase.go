package firebase

import (
	"context"

	"bitbucket.org/pensarmais/cycleforlisbon/src/database/models"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	Fcm *messaging.Client
}

// Initializes the Firebase App and Messaging Client.
func New(credentialsFile *string) (*Client, error) {
	opt := option.WithCredentialsFile(*credentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	msg, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{msg}, nil
}

// Creates a test message to be used in a multicast dry-run, to test token
// validity. It's not meant to be sent to the devices.
func NewTestMessage(tokens []string) messaging.MulticastMessage {
	return messaging.MulticastMessage{
		Tokens: tokens,
		Data: map[string]string{
			"test": "true",
		},
	}
}

func NewAchievementMessage(
	token string,
	achievement models.Achievement,
	host string,
) messaging.Message {
	return messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title:    achievement.Name,
			Body:     achievement.Desc,
			ImageURL: host + achievement.ImageURI,
		},
		Data: map[string]string{
			"type": "achievement",
			"code": achievement.Code,
		},
	}
}

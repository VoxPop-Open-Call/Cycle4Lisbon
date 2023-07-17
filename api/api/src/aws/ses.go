package aws

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SES embeds the ses.Client.
type SES struct {
	*ses.Client

	fromName    string
	domain      string
	apiScheme   string
	apiEndpoint string
	source      string
	templates   map[string]*template.Template
}

// HTML template filenames.
const (
	base            = "resources/templates/base.html"
	passwordReset   = "resources/templates/password-reset.html"
	passwordChanged = "resources/templates/password-changed.html"
)

func newSES(config aws.Config, domain, fromName, apiScheme, apiEndpoint string) *SES {
	return &SES{
		Client:      ses.NewFromConfig(config),
		fromName:    fromName,
		domain:      domain,
		apiScheme:   apiScheme,
		apiEndpoint: apiEndpoint,
		source:      fromName + " <noreply@" + domain + ">",
		templates:   parseTemplates(),
	}
}

func parseTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template)

	for _, src := range []string{
		passwordReset,
		passwordChanged,
	} {
		result[src] = template.Must(template.ParseFiles(src, base))
	}

	return result
}

type sendEmailParams struct {
	Destination string
	Subject     string
	Body        string
}

// sendEmail composes an email and queues it for sending with the ses client.
func (c *SES) sendEmail(params sendEmailParams) error {
	_, err := c.Client.SendEmail(context.TODO(), &ses.SendEmailInput{
		Source: aws.String(c.source),
		Destination: &types.Destination{
			ToAddresses: []string{params.Destination},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: &params.Subject,
			},
			Body: &types.Body{
				Html: &types.Content{
					Data:    &params.Body,
					Charset: aws.String("UTF-8"),
				},
			},
		},
	})

	return err
}

func (c *SES) sendTemplatedEmail(
	subject, destination, template string,
	data any,
) error {
	var body bytes.Buffer
	err := c.templates[template].ExecuteTemplate(&body, "base", data)
	if err != nil {
		return err
	}

	return c.sendEmail(sendEmailParams{
		Destination: destination,
		Subject:     c.fromName + " " + subject,
		Body:        string(body.Bytes()),
	})
}

func (c *SES) resetPasswordLink(email, code string) string {
	return fmt.Sprintf("%s://%s/password/redirect-reset?email=%s&code=%s",
		c.apiScheme, c.apiEndpoint, url.QueryEscape(email), code)
}

func (c *SES) SendPasswordResetEmail(email, code string) error {
	return c.sendTemplatedEmail("Password Reset", email, passwordReset,
		struct{ Link string }{c.resetPasswordLink(email, code)})
}

func (c *SES) SendPasswordChangedEmail(email string) error {
	return c.sendTemplatedEmail("Password Changed", email, passwordChanged,
		struct{}{})
}

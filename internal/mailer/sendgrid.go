package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGridMailer(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

// Sends an email using SendGrid with the specified template and data.
// Data is used to populate the template.
func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	fmt.Printf("Sending email to %s <%s> from %s using template %s\n", username, email, m.fromEmail, templateFile)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("failed to execute subject template: %w", err)
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return fmt.Errorf("failed to execute body template: %w", err)
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{Enable: &isSandbox},
	})

	for i := 0; i < MAX_RETRIES; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %v, attempt %d/%d: %v", email, i+1, MAX_RETRIES, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Printf("Email sent to %v with status code %d", email, response.StatusCode)
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return nil
		} else {
			return fmt.Errorf("failed to sent email to %v with status code %d", email, response.StatusCode)
		}
	}

	return fmt.Errorf("failed to send email to %v after %d attempts", email, MAX_RETRIES)
}

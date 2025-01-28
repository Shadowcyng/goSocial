package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"time"

	gomail "gopkg.in/mail.v2"
)

type MailerService struct {
	fromEmail string
	apiKey    string
}

func NewMailerService(apiKey, fromEmail string) (*MailerService, error) {
	if apiKey == "" {
		return &MailerService{}, errors.New("api key is required")
	}

	return &MailerService{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}, nil
}

func (m *MailerService) Send(templateFile string, username string, email string, data any, isSandbox bool) error {
	// template parsing and building
	tmpl, err := template.ParseFS(FS, fmt.Sprintf("templates/%s", templateFile))
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}
	message := gomail.NewMessage()

	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("subject", subject.String())
	message.AddAlternative("text/html", body.String())

	dialer := gomail.NewDialer("sandbox.smtp.mailtrap.io", 2525, "6e4b23f56d5e95", m.apiKey)
	if err := dialer.DialAndSend(message); err != nil {
		var retryErr error
		for i := 0; i < maxRetries; i++ {
			retryErr = dialer.DialAndSend(message)
			if err != nil {
				log.Printf("Filed to send email to %v, attemp %d of %d", email, i+1, maxRetries)
				// exponential backoff
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
			log.Printf("Email successfully send to %s ", email)
			return nil
		}
		return fmt.Errorf("failed to send email after %d attempts, error: %v", maxRetries, retryErr)
	}
	return nil
}

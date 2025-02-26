package email

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// sendWithSendgrid sends an email using the Sendgrid API
func (s *Service) sendWithSendgrid(data EmailData, htmlContent, textContent string) error {
	from := mail.NewEmail(data.FromName, data.From)
	to := mail.NewEmail("", data.To)
	message := mail.NewSingleEmail(from, data.Subject, to, textContent, htmlContent)

	response, err := s.sendgridClient.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email via Sendgrid: %w", err)
	}

	if response.StatusCode != 202 {
		return fmt.Errorf("unexpected Sendgrid status code: %d, body: %s", response.StatusCode, response.Body)
	}

	return nil
}

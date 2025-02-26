package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"time"
)

// sendWithSMTP sends an email using SMTP
func (s *Service) sendWithSMTP(data EmailData, htmlContent, textContent string) error {

	if s.config.Sendgrid.APIKey != "" {
		return s.sendWithSendgrid(data, htmlContent, textContent)
	}

	config := s.config.SMTP[string(s.provider)]

	// Create buffer for the message
	var buf bytes.Buffer

	// Write headers
	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", data.FromName, data.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", data.To))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Generate a unique boundary
	boundary := fmt.Sprintf("_MULTIPART_MIXED_BOUNDARY_%d", time.Now().UnixNano())
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))

	// Write the plaintext part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(textContent)))
	buf.WriteString("\r\n")

	// Write the HTML part
	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(htmlContent)))
	buf.WriteString("\r\n")

	// Close the multipart message
	buf.WriteString(fmt.Sprintf("\r\n--%s--", boundary))

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	if err := smtp.SendMail(addr, auth, data.From, []string{data.To}, buf.Bytes()); err != nil {
		return fmt.Errorf("sending email via SMTP: %w", err)
	}

	return nil
}

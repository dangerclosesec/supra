//go:build ignore

package mailer

import (
	"context"
	"fmt"
)

// Mailer defines the interface for sending emails
type Mailer interface {
	Send(ctx context.Context, email Email) error
}

// Email represents the email message structure
type Email struct {
	From        string
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Content  []byte
	MIMEType string
}

// Provider represents an email service provider
type Provider interface {
	SendEmail(ctx context.Context, email Email) error
}

// BaseMailer provides common functionality for all mailers
type BaseMailer struct {
	provider Provider
	from     string
}

// NewBaseMailer creates a new base mailer instance
func NewBaseMailer(provider Provider, defaultFrom string) BaseMailer {
	return BaseMailer{
		provider: provider,
		from:     defaultFrom,
	}
}

// UserMailer handles user-related email communications
type UserMailer struct {
	BaseMailer
	templates Templates
}

// Templates interface defines methods for getting email templates
type Templates interface {
	GetTemplate(name string) (string, error)
	GetHTMLTemplate(name string) (string, error)
}

// NewUserMailer creates a new user mailer instance
func NewUserMailer(provider Provider, templates Templates, defaultFrom string) *UserMailer {
	return &UserMailer{
		BaseMailer: NewBaseMailer(provider, defaultFrom),
		templates:  templates,
	}
}

// SendWelcome sends a welcome email to new users
func (m *UserMailer) SendWelcome(ctx context.Context, user User) error {
	template, err := m.templates.GetTemplate("welcome")
	if err != nil {
		return fmt.Errorf("getting welcome template: %w", err)
	}

	htmlTemplate, err := m.templates.GetHTMLTemplate("welcome")
	if err != nil {
		return fmt.Errorf("getting welcome HTML template: %w", err)
	}

	email := Email{
		From:     m.from,
		To:       []string{user.Email},
		Subject:  "Welcome to Our Service",
		Body:     fmt.Sprintf(template, user.Name),
		HTMLBody: fmt.Sprintf(htmlTemplate, user.Name),
	}

	return m.provider.SendEmail(ctx, email)
}

// SendPasswordReset sends password reset instructions
func (m *UserMailer) SendPasswordReset(ctx context.Context, user User, token string) error {
	template, err := m.templates.GetTemplate("password_reset")
	if err != nil {
		return fmt.Errorf("getting password reset template: %w", err)
	}

	htmlTemplate, err := m.templates.GetHTMLTemplate("password_reset")
	if err != nil {
		return fmt.Errorf("getting password reset HTML template: %w", err)
	}

	email := Email{
		From:     m.from,
		To:       []string{user.Email},
		Subject:  "Password Reset Instructions",
		Body:     fmt.Sprintf(template, user.Name, token),
		HTMLBody: fmt.Sprintf(htmlTemplate, user.Name, token),
	}

	return m.provider.SendEmail(ctx, email)
}

// User represents a user in the system
type User struct {
	Email string
	Name  string
}

// Example provider implementations:

// SMTPProvider implements Provider interface for SMTP servers
type SMTPProvider struct {
	host     string
	port     int
	username string
	password string
}

// SendEmail sends an email using SMTP
func (p *SMTPProvider) SendEmail(ctx context.Context, email Email) error {
	// SMTP implementation
	return nil
}

// SendGridProvider implements Provider interface for SendGrid
type SendGridProvider struct {
	apiKey string
}

// SendEmail sends an email using SendGrid
func (p *SendGridProvider) SendEmail(ctx context.Context, email Email) error {
	// SendGrid implementation
	return nil
}

// Usage example:
func Example() {
	// Initialize templates
	templates := &MyTemplates{}

	// Initialize provider
	provider := &SMTPProvider{
		host:     "smtp.example.com",
		port:     587,
		username: "user",
		password: "pass",
	}

	// Create mailer
	userMailer := NewUserMailer(provider, templates, "noreply@example.com")

	// Send welcome email
	user := User{
		Email: "user@example.com",
		Name:  "John Doe",
	}

	ctx := context.Background()
	if err := userMailer.SendWelcome(ctx, user); err != nil {
		// Handle error
	}
}

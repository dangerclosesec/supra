// internal/email/service.go
package email

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/dangerclosesec/supra"
	"github.com/dangerclosesec/supra/internal/config"
	"github.com/sendgrid/sendgrid-go"
)

var templateFS = supra.EmailFS

// Provider identifies supported email providers
type Provider string

const (
	ProviderSMTP     Provider = "smtp"
	ProviderSendgrid Provider = "sendgrid"

	DefaultTemplatePath = "templates/emails"
)

// EmailData contains all necessary information for sending an email
type EmailData struct {
	To           string
	From         string
	FromName     string
	Subject      string
	TemplateName string
	TemplateData interface{}
}

// Service handles email operations
type Service struct {
	config         *config.Config
	provider       Provider
	sendgridClient *sendgrid.Client
	Templates      map[string]*Template
}

type Template struct {
	HTML      *template.Template
	Plaintext *template.Template
}

// NewEmailService creates a new email service instance
func NewEmailService(config *config.Config, provider Provider) (*Service, error) {
	s := &Service{
		config:    config,
		provider:  provider,
		Templates: make(map[string]*Template),
	}

	if provider == ProviderSendgrid {
		s.sendgridClient = sendgrid.NewSendClient(config.Sendgrid.APIKey)
	}

	if err := s.loadTemplates(); err != nil {
		return nil, fmt.Errorf("loading email templates: %w", err)
	}

	return s, nil
}

// loadTemplates loads all email templates from the embedded filesystem
func (s *Service) loadTemplates() error {
	templateGroups, err := templateFS.ReadDir(DefaultTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read email templates directory: %w", err)
	}

	if len(templateGroups) == 0 {
		return fmt.Errorf("no email templates found")
	}

	for _, group := range templateGroups {
		if !group.IsDir() {
			continue
		}

		// fmt.Printf("Group: %s\n", group.Name())

		groupPath := DefaultTemplatePath + "/" + group.Name()
		groupEntries, err := templateFS.ReadDir(groupPath)
		if err != nil {
			return fmt.Errorf("failed to read email template group %s: %w", group.Name(), err)
		}

		if len(groupEntries) != 2 {
			return fmt.Errorf("invalid email template group %s: must contain exactly two files (HTML and plaintext)", group.Name())
		}

		tmpl := Template{
			HTML:      template.Must(template.ParseFS(templateFS, groupPath+"/html.tmpl")),
			Plaintext: template.Must(template.ParseFS(templateFS, groupPath+"/plaintext.tmpl")),
		}

		s.Templates[group.Name()] = &tmpl
	}

	// fmt.Printf("Loaded %d templates\n", len(s.Templates))

	return nil
}

// SendEmail sends an email using the configured provider
func (s *Service) SendEmail(data EmailData) error {
	// Renders both HTML and text versions of the email
	htmlContent, textContent, err := s.renderTemplate(data.TemplateName, data.TemplateData)
	if err != nil {
		return fmt.Errorf("rendering HTML template: %w", err)
	}

	switch s.provider {
	case ProviderSendgrid:
		if data.From == "" {
			data.From = s.config.Sendgrid.From
		}
		return s.sendWithSendgrid(data, htmlContent, textContent)
	case ProviderSMTP:
		if data.From == "" {
			return fmt.Errorf("missing sender email address (From)")
		}
		return s.sendWithSMTP(data, htmlContent, textContent)
	default:
		return fmt.Errorf("unsupported email provider: %s", s.provider)
	}
}

// renderTemplate renders a template with the given data
func (s *Service) renderTemplate(name string, data interface{}) (string, string, error) {
	tmpl, exists := s.Templates[name]
	if !exists {
		return "", "", fmt.Errorf("template %s not found", name)
	}

	var htmlbuf bytes.Buffer
	if err := (*tmpl).HTML.Execute(&htmlbuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	var textbuf bytes.Buffer
	if err := (*tmpl).Plaintext.Execute(&textbuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	return htmlbuf.String(), textbuf.String(), nil
}

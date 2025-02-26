// internal/email/mailers/new_account_verification.go
package mailer

import "github.com/dangerclosesec/supra/internal/email"

// VerificationTemplateData contains data for the verification email template
type VerificationTemplateData struct {
	FirstName        string
	VerificationLink string
}

// SendVerificationEmail sends a verification email to the user
func SendVerificationEmail(s *email.Service, to, firstName, verificationLink string) error {
	templateData := VerificationTemplateData{
		FirstName:        firstName,
		VerificationLink: verificationLink,
	}

	fromName := "RocketBox"

	emailData := email.EmailData{
		To:           to,
		FromName:     fromName,
		Subject:      "Welcome to RocketBox! Please verify your email",
		TemplateName: "new_account_verification",
		TemplateData: templateData,
	}

	return s.SendEmail(emailData)
}

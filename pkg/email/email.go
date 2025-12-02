package email

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(toEmail, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// If SMTP config is missing, just log it (for dev/testing without real SMTP)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	if smtpHost == "" || smtpUser == "" {
		fmt.Printf("Mock Email to %s: Verify at %s/verify-email?token=%s\n", toEmail, baseURL, token)
		return nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	link := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Verify your email\r\n"+
		"\r\n"+
		"Please click the link below to verify your email address:\r\n"+
		"%s\r\n", toEmail, link))

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
	if err != nil {
		return err
	}
	return nil
}

func SendNotificationEmail(toEmail, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	
	// If SMTP config is missing, just log it
	if smtpHost == "" || smtpUser == "" {
		fmt.Printf("Mock Email to %s: Subject: %s\nBody: %s\n", toEmail, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", toEmail, subject, body))

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, msg)
	if err != nil {
		return err
	}
	return nil
}

// Package email provides email sending functionality using SMTP.
package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// Service handles email operations using SMTP
type Service struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	senderEmail  string
	senderName   string
	appURL       string
}

// NewService creates a new email service instance
func NewService(smtpHost, smtpPort, smtpUsername, smtpPassword, senderEmail, senderName, appURL string) (*Service, error) {
	return &Service{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		senderEmail:  senderEmail,
		senderName:   senderName,
		appURL:       appURL,
	}, nil
}

// SendVerificationEmail sends an email verification link to the user
func (s *Service) SendVerificationEmail(ctx context.Context, recipientEmail, recipientName, verificationToken string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, verificationToken)

	subject := "Verify Your Email - FacturMe"
	htmlBody := s.getVerificationEmailHTML(recipientName, verificationURL)
	textBody := s.getVerificationEmailText(recipientName, verificationURL)

	// Build email message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.senderName, s.senderEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipientEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary-string\"\r\n")
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString("--boundary-string\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString("--boundary-string\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString("--boundary-string--")

	// Set up authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Connect to the SMTP server with TLS
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: s.smtpHost,
	}

	// Dial with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender
	if err = client.Mail(s.senderEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(recipientEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	err = client.Quit()
	if err != nil {
		return fmt.Errorf("failed to quit: %w", err)
	}

	return nil
}

// getVerificationEmailHTML returns the HTML template for verification email
func (s *Service) getVerificationEmailHTML(name, verificationURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0f172a;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse; background-color: #0f172a;">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" style="width: 100%%; max-width: 600px; border-collapse: collapse; background-color: #1e293b; border-radius: 16px; overflow: hidden; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3);">
                    <!-- Header -->
                    <tr>
                        <td align="center" style="padding: 40px 40px 30px 40px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 700;">FacturMe</h1>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #f1f5f9; font-size: 24px; font-weight: 600;">Hi %s! 👋</h2>
                            <p style="margin: 0 0 20px 0; color: #cbd5e1; font-size: 16px; line-height: 1.6;">
                                Welcome to FacturMe! We're excited to have you on board. To get started, please verify your email address by clicking the button below.
                            </p>

                            <!-- Button -->
                            <table role="presentation" style="margin: 30px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="%s" style="display: inline-block; padding: 16px 32px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 16px;">
                                            Verify Email Address
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 30px 0 10px 0; color: #cbd5e1; font-size: 14px; line-height: 1.6;">
                                Or copy and paste this link into your browser:
                            </p>
                            <p style="margin: 0; padding: 12px; background-color: #334155; border-radius: 6px; color: #94a3b8; font-size: 13px; word-break: break-all;">
                                %s
                            </p>

                            <p style="margin: 30px 0 0 0; color: #94a3b8; font-size: 14px; line-height: 1.6;">
                                This link will expire in <strong>24 hours</strong>.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 30px 40px; background-color: #0f172a; border-top: 1px solid #334155;">
                            <p style="margin: 0 0 10px 0; color: #64748b; font-size: 12px; line-height: 1.5;">
                                If you didn't create an account with FacturMe, you can safely ignore this email.
                            </p>
                            <p style="margin: 0; color: #64748b; font-size: 12px;">
                                © 2025 FacturMe. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name, verificationURL, verificationURL)
}

// getVerificationEmailText returns the plain text template for verification email
func (s *Service) getVerificationEmailText(name, verificationURL string) string {
	return fmt.Sprintf(`
Hi %s!

Welcome to FacturMe! We're excited to have you on board.

To get started, please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you didn't create an account with FacturMe, you can safely ignore this email.

© 2025 FacturMe. All rights reserved.
`, name, verificationURL)
}

// SendPasswordResetEmail sends a password reset link to the user
func (s *Service) SendPasswordResetEmail(ctx context.Context, recipientEmail, recipientName, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, resetToken)

	subject := "Reset Your Password - FacturMe"
	htmlBody := s.getPasswordResetEmailHTML(recipientName, resetURL)
	textBody := s.getPasswordResetEmailText(recipientName, resetURL)

	// Build email message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.senderName, s.senderEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipientEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary-string\"\r\n")
	msg.WriteString("\r\n")

	// Plain text part
	msg.WriteString("--boundary-string\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString("--boundary-string\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString("--boundary-string--")

	// Set up authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Connect to the SMTP server with TLS
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: s.smtpHost,
	}

	// Dial with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender
	if err = client.Mail(s.senderEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(recipientEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	err = client.Quit()
	if err != nil {
		return fmt.Errorf("failed to quit: %w", err)
	}

	return nil
}

// getPasswordResetEmailHTML returns the HTML template for password reset email
func (s *Service) getPasswordResetEmailHTML(name, resetURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0f172a;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse; background-color: #0f172a;">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" style="width: 100%%; max-width: 600px; border-collapse: collapse; background-color: #1e293b; border-radius: 16px; overflow: hidden; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3);">
                    <!-- Header -->
                    <tr>
                        <td align="center" style="padding: 40px 40px 30px 40px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: 700;">FacturMe</h1>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #f1f5f9; font-size: 24px; font-weight: 600;">Hi %s! 👋</h2>
                            <p style="margin: 0 0 20px 0; color: #cbd5e1; font-size: 16px; line-height: 1.6;">
                                We received a request to reset your password for your FacturMe account. Click the button below to create a new password.
                            </p>

                            <!-- Button -->
                            <table role="presentation" style="margin: 30px 0;">
                                <tr>
                                    <td align="center">
                                        <a href="%s" style="display: inline-block; padding: 16px 32px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600; font-size: 16px;">
                                            Reset Password
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 30px 0 10px 0; color: #cbd5e1; font-size: 14px; line-height: 1.6;">
                                Or copy and paste this link into your browser:
                            </p>
                            <p style="margin: 0; padding: 12px; background-color: #334155; border-radius: 6px; color: #94a3b8; font-size: 13px; word-break: break-all;">
                                %s
                            </p>

                            <p style="margin: 30px 0 0 0; color: #94a3b8; font-size: 14px; line-height: 1.6;">
                                This link will expire in <strong>1 hour</strong>.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 30px 40px; background-color: #0f172a; border-top: 1px solid #334155;">
                            <p style="margin: 0 0 10px 0; color: #64748b; font-size: 12px; line-height: 1.5;">
                                If you didn't request a password reset, you can safely ignore this email. Your password will not be changed.
                            </p>
                            <p style="margin: 0; color: #64748b; font-size: 12px;">
                                © 2025 FacturMe. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name, resetURL, resetURL)
}

// getPasswordResetEmailText returns the plain text template for password reset email
func (s *Service) getPasswordResetEmailText(name, resetURL string) string {
	return fmt.Sprintf(`
Hi %s!

We received a request to reset your password for your FacturMe account.

To reset your password, click the link below:

%s

This link will expire in 1 hour.

If you didn't request a password reset, you can safely ignore this email. Your password will not be changed.

© 2025 FacturMe. All rights reserved.
`, name, resetURL)
}

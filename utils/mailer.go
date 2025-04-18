package utils

import (
	"net/smtp"
	"os"
	"urllite/types"
)

type mailer struct {
	auth        smtp.Auth
	smtpHost    string
	smtpPort    string
	mailerEmail string
}

type Mailer interface {
	SendOtpForEmailVerification(user *types.User, otp *types.Otp) error
}

func NewMailer() Mailer {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("MAILER_EMAIL")
	password := os.Getenv("MAILER_PASSWORD")
	return &mailer{auth: smtp.PlainAuth("", from, password, smtpHost), smtpHost: smtpHost, smtpPort: smtpPort, mailerEmail: from}
}

func (m *mailer) SendOtpForEmailVerification(user *types.User, otp *types.Otp) error {
	subject := "Email Verification OTP"
	body := "Dear " + user.Name + ", Your otp is: " + otp.Otp + ". This otp is valid only for 10 minutes."
	message := subject + "\r\n\r\n" + body
	err := smtp.SendMail(m.smtpHost+":"+m.smtpPort, m.auth, m.mailerEmail, []string{user.Email}, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

package service

import (
	"fmt"
	"mentalchat/internal/config"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	cfg         *config.EmailConfig
	frontendURL string
}

func NewEmailService(cfg *config.EmailConfig) *EmailService {
	return &EmailService{
		cfg:         cfg,
		frontendURL: cfg.FromEmail, // Will be set from app config
	}
}

func (s *EmailService) SetFrontendURL(url string) {
	s.frontendURL = url
}

func (s *EmailService) SendVerificationEmail(email, verificationURL string) error {
	// Send email
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.FromName+" <"+s.cfg.FromEmail+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Подтверждение email - MentalChat")
	m.SetBody("text/html", s.buildVerificationEmailHTML(verificationURL))

	if err := s.sendEmail(m); err != nil {
		return err
	}

	return nil
}

func (s *EmailService) SendPasswordResetEmail(email, resetURL string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.FromName+" <"+s.cfg.FromEmail+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Сброс пароля - MentalChat")
	m.SetBody("text/html", s.buildPasswordResetEmailHTML(resetURL))

	if err := s.sendEmail(m); err != nil {
		return err
	}

	return nil
}

func (s *EmailService) SendMarketingEmail(email, subject, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.FromName+" <"+s.cfg.FromEmail+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	if err := s.sendEmail(m); err != nil {
		return err
	}

	return nil
}

func (s *EmailService) buildVerificationEmailHTML(verificationURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Подтверждение email - MentalChat</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%); padding: 30px; border-radius: 15px; text-align: center;">
        <h1 style="color: #4a4a4a;">Добро пожаловать в MentalChat! 🌸</h1>
        <p>Пожалуйста, подтвердите свой email, чтобы начать пользоваться нашим сервисом.</p>
        <a href="%s" style="display: inline-block; background: #d4a373; color: white; padding: 12px 30px; text-decoration: none; border-radius: 25px; margin: 20px 0;">Подтвердить email</a>
        <p style="font-size: 14px; color: #666;">Если кнопка не работает, скопируйте и вставьте этот URL в браузер:<br>%s</p>
    </div>
    <div style="text-align: center; padding: 20px; color: #666;">
        <p>С любовью,<br>MentalChat Team</p>
    </div>
</body>
</html>
`, verificationURL, verificationURL)
}

func (s *EmailService) buildPasswordResetEmailHTML(resetURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Сброс пароля - MentalChat</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%); padding: 30px; border-radius: 15px; text-align: center;">
        <h1 style="color: #4a4a4a;">Сброс пароля</h1>
        <p>Вы запросили сброс пароля. Нажмите кнопку ниже, чтобы создать новый пароль.</p>
        <a href="%s" style="display: inline-block; background: #d4a373; color: white; padding: 12px 30px; text-decoration: none; border-radius: 25px; margin: 20px 0;">Сбросить пароль</a>
        <p style="font-size: 14px; color: #666;">Ссылка действительна в течение 1 часа.</p>
    </div>
    <div style="text-align: center; padding: 20px; color: #666;">
        <p>С любовью,<br>MentalChat Team</p>
    </div>
</body>
</html>
`, resetURL)
}

func (s *EmailService) GetFrontendURL() string {
	return s.frontendURL
}

func (s *EmailService) sendEmail(m *gomail.Message) error {
	smtpPort, err := strconv.Atoi(s.cfg.SMTPPort)
	if err != nil {
		smtpPort = 587 // Default port
	}

	d := gomail.NewDialer(s.cfg.SMTPHost, smtpPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)

	// Send the email
	return d.DialAndSend(m)
}

func (s *EmailService) QueueMarketingEmails() error {
	// Implementation for queuing marketing emails
	return nil
}

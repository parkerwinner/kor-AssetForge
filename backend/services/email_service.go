package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type EmailProvider string

const (
	ProviderSendGrid EmailProvider = "sendgrid"
	ProviderSES      EmailProvider = "ses"
)

const (
	defaultFromName           = "AssetForge"
	defaultFromAddress        = "no-reply@assetforge.io"
	defaultVerificationURL    = "https://app.assetforge.io/verify-email"
	emailQueueBuffer          = 100
	emailContentBoundary      = "--assetforge-boundary"
	emailContentTypeHTML      = "text/html; charset=UTF-8"
	emailContentTypePlainText = "text/plain; charset=UTF-8"
)

type EmailService interface {
	SendVerificationEmail(toEmail, toName, verificationToken string) error
	SendKYCStatusUpdate(toEmail, toName, status, reviewNotes string) error
	SendTransactionConfirmation(toEmail, toName, txHash string, amount int64, assetID uint, fromAddress, toAddress string) error
	SendApprovalPendingEmail(toEmail, toName string, requestID uint, expiresAt time.Time) error
	SendScheduledReportEmail(recipients []string, reportName, format, fileName, body string) error
	SendGDPRDataExportLink(toEmail, toName, downloadLink string) error
}

func (s *emailService) SendApprovalPendingEmail(toEmail, toName string, requestID uint, expiresAt time.Time) error {
	subject := "Asset transfer approval required"
	plain := fmt.Sprintf("Hi %s,\n\nApproval request #%d is waiting for your decision. It expires at %s.\n", toName, requestID, expiresAt.Format(time.RFC3339))
	html := fmt.Sprintf("<p>Hi %s,</p><p>Approval request <strong>#%d</strong> is waiting for your decision.</p><p>It expires at %s.</p>", toName, requestID, expiresAt.Format(time.RFC3339))
	return s.queueEmail(&EmailMessage{To: toEmail, ToName: toName, Subject: subject, PlainText: plain, HTML: html})
}

type emailService struct {
	provider        EmailProvider
	fromAddress     string
	fromName        string
	verificationURL string
	sendGridAPIKey  string
	sesRegion       string
	sesSMTPUsername string
	sesSMTPPassword string
	queue           chan *EmailMessage
}

type EmailMessage struct {
	To        string
	ToName    string
	Subject   string
	PlainText string
	HTML      string
}

func NewEmailServiceFromEnv() EmailService {
	provider := EmailProvider(strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER"))))
	if provider == "" {
		if os.Getenv("SENDGRID_API_KEY") != "" {
			provider = ProviderSendGrid
		} else if os.Getenv("SES_SMTP_USERNAME") != "" && os.Getenv("SES_SMTP_PASSWORD") != "" {
			provider = ProviderSES
		}
	}

	fromAddress := os.Getenv("EMAIL_FROM_ADDRESS")
	if fromAddress == "" {
		fromAddress = defaultFromAddress
	}
	fromName := os.Getenv("EMAIL_FROM_NAME")
	if fromName == "" {
		fromName = defaultFromName
	}
	verificationURL := os.Getenv("EMAIL_VERIFICATION_URL_BASE")
	if verificationURL == "" {
		verificationURL = defaultVerificationURL
	}

	emailSvc := &emailService{
		provider:        provider,
		fromAddress:     fromAddress,
		fromName:        fromName,
		verificationURL: verificationURL,
		sendGridAPIKey:  os.Getenv("SENDGRID_API_KEY"),
		sesRegion:       os.Getenv("SES_REGION"),
		sesSMTPUsername: os.Getenv("SES_SMTP_USERNAME"),
		sesSMTPPassword: os.Getenv("SES_SMTP_PASSWORD"),
		queue:           make(chan *EmailMessage, emailQueueBuffer),
	}

	go emailSvc.worker()
	return emailSvc
}

func (s *emailService) queueEmail(msg *EmailMessage) error {
	select {
	case s.queue <- msg:
		return nil
	default:
		return errors.New("email queue is full")
	}
}

func (s *emailService) worker() {
	for msg := range s.queue {
		var err error
		switch s.provider {
		case ProviderSendGrid:
			err = s.sendWithSendGrid(msg)
		case ProviderSES:
			err = s.sendWithSES(msg)
		default:
			err = s.sendWithSMTP(msg)
		}
		if err != nil {
			log.Printf("email_service: failed to send email to %s: %v", msg.To, err)
		}
	}
}

func (s *emailService) sendWithSMTP(msg *EmailMessage) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpHost == "" || smtpPort == "" {
		log.Printf("email_service: SMTP_HOST/SMTP_PORT not configured; writing email to log:\nTo: %s\nSubject: %s\nBody: %s", msg.To, msg.Subject, msg.PlainText)
		return nil
	}

	auth := smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), smtpHost)
	body := fmt.Sprintf("MIME-Version: 1.0\r\nContent-Type: %s; boundary=%s\r\nSubject: %s\r\nTo: %s\r\nFrom: %s <%s>\r\n\r\n", "multipart/alternative", emailContentBoundary, msg.Subject, msg.To, s.fromName, s.fromAddress)
	body += fmt.Sprintf("--%s\r\nContent-Type: %s\r\n\r\n%s\r\n", emailContentBoundary, emailContentTypePlainText, msg.PlainText)
	body += fmt.Sprintf("--%s\r\nContent-Type: %s\r\n\r\n%s\r\n", emailContentBoundary, emailContentTypeHTML, msg.HTML)
	body += fmt.Sprintf("--%s--", emailContentBoundary)

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, s.fromAddress, []string{msg.To}, []byte(body))
}

func (s *emailService) sendWithSendGrid(msg *EmailMessage) error {
	if s.sendGridAPIKey == "" {
		return errors.New("SendGrid API key not configured")
	}
	url := "https://api.sendgrid.com/v3/mail/send"
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]string{
					{"email": msg.To, "name": msg.ToName},
				},
			},
		},
		"from":             map[string]string{"email": s.fromAddress, "name": s.fromName},
		"subject":          msg.Subject,
		"content": []map[string]string{
			{"type": "text/plain", "value": msg.PlainText},
			{"type": "text/html", "value": msg.HTML},
		},
	}
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+s.sendGridAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("SendGrid API returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *emailService) sendWithSES(msg *EmailMessage) error {
	return errors.New("AWS SES email integration not implemented in this version")
}

func (s *emailService) SendVerificationEmail(toEmail, toName, verificationToken string) error {
	link := fmt.Sprintf("%s?token=%s", s.verificationURL, verificationToken)
	subject := "Verify your email address"
	plain := fmt.Sprintf("Hi %s,\n\nPlease verify your email by clicking: %s\n", toName, link)
	html := fmt.Sprintf("<p>Hi %s,</p><p>Please verify your email by clicking <a href=\"%s\">here</a>.</p>", toName, link)
	return s.queueEmail(&EmailMessage{To: toEmail, ToName: toName, Subject: subject, PlainText: plain, HTML: html})
}

func (s *emailService) SendKYCStatusUpdate(toEmail, toName, status, reviewNotes string) error {
	subject := fmt.Sprintf("KYC status update: %s", strings.Title(status))
	plain := fmt.Sprintf("Hi %s,\n\nYour KYC status has been updated to %s.\n\nReview notes: %s\n\nIf you have questions, contact support.\n", toName, strings.Title(status), reviewNotes)
	html := fmt.Sprintf("<p>Hi %s,</p><p>Your KYC status has been updated to <strong>%s</strong>.</p><p><strong>Review notes:</strong> %s</p><p>If you have questions, reply to this email or contact support.</p>", toName, strings.Title(status), reviewNotes)
	return s.queueEmail(&EmailMessage{To: toEmail, ToName: toName, Subject: subject, PlainText: plain, HTML: html})
}

func (s *emailService) SendTransactionConfirmation(toEmail, toName, txHash string, amount int64, assetID uint, fromAddress, toAddress string) error {
	subject := "Transaction confirmation from AssetForge"
	plain := fmt.Sprintf("Hi %s,\n\nYour transaction has been recorded successfully.\n\nTransaction hash: %s\nAsset ID: %d\nAmount: %d\nFrom: %s\nTo: %s\n\nThank you for using AssetForge.\n", toName, txHash, assetID, amount, fromAddress, toAddress)
	html := fmt.Sprintf("<p>Hi %s,</p><p>Your transaction has been recorded successfully.</p><ul><li><strong>Transaction hash:</strong> %s</li><li><strong>Asset ID:</strong> %d</li><li><strong>Amount:</strong> %d</li><li><strong>From:</strong> %s</li><li><strong>To:</strong> %s</li></ul><p>Thank you for using AssetForge.</p>", toName, txHash, assetID, amount, fromAddress, toAddress)
	return s.queueEmail(&EmailMessage{To: toEmail, ToName: toName, Subject: subject, PlainText: plain, HTML: html})
}

func (s *emailService) SendScheduledReportEmail(recipients []string, reportName, format, fileName, body string) error {
	if len(recipients) == 0 {
		return errors.New("at least one recipient is required")
	}
	subject := fmt.Sprintf("Scheduled report ready: %s", reportName)
	plain := fmt.Sprintf("Your scheduled %s report is ready.\n\nFile: %s\nFormat: %s\n\n%s\n", reportName, fileName, strings.ToUpper(format), body)
	html := fmt.Sprintf("<p>Your scheduled <strong>%s</strong> report is ready.</p><p><strong>File:</strong> %s<br><strong>Format:</strong> %s</p><pre style=\"white-space:pre-wrap\">%s</pre>", reportName, fileName, strings.ToUpper(format), body)
	for _, recipient := range recipients {
		if err := s.queueEmail(&EmailMessage{To: recipient, Subject: subject, PlainText: plain, HTML: html}); err != nil {
			return err
		}
	}
	return nil
}

func (s *emailService) SendGDPRDataExportLink(toEmail, toName, downloadLink string) error {
	subject := "Your GDPR Data Export is Ready"
	plain := fmt.Sprintf("Hi %s,\n\nYour GDPR data export request is ready for download. Please click the link below to retrieve your file:\n\n%s\n\nNote: This link will expire in 7 days.\n\nThank you,\nAssetForge Team\n", toName, downloadLink)
	html := fmt.Sprintf("<p>Hi %s,</p><p>Your GDPR data export request is ready for download. Please click the link below to retrieve your file:</p><p><a href=\"%s\" style=\"display:inline-block;padding:12px 20px;background:#1a73e8;color:#fff;text-decoration:none;border-radius:4px;\">Download Data Export</a></p><p><em>Note: This link will expire in 7 days.</em></p><p>Thank you,<br/>AssetForge Team</p>", toName, downloadLink)
	return s.queueEmail(&EmailMessage{To: toEmail, ToName: toName, Subject: subject, PlainText: plain, HTML: html})
}

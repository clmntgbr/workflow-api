package mail

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"
)

var _ port.MailSender = (*Sender)(nil)

//go:embed templates/*.html
var templateFS embed.FS

var templateFiles = []string{
	"workflow_finished_success",
	"workflow_finished_failed",
	"workflow_finished_cancelled",
	"invoice_payment_succeeded",
	"invoice_payment_failed",
	"subscription_plan_changed",
	"subscription_renewal_upcoming",
	"payment_method_expiring",
	"quota_threshold_reached",
	"quota_exceeded",
	"run_export_ready",
	"run_export_failed",
}

const defaultMailFrom = "Workflow <noreply@localhost>"

type smtpConfig struct {
	addr string
	auth smtp.Auth
}

type smtpSendFunc func(ctx context.Context, cfg smtpConfig, from string, to []string, msg []byte) error

type Sender struct {
	from           string
	logoURL        string
	companyAddress string
	preferencesURL string
	smtp           *smtpConfig
	templates      map[string]*template.Template
	send           smtpSendFunc
}

func NewSender(cfg *config.Config) (*Sender, error) {
	templates, err := parseTemplates()
	if err != nil {
		return nil, err
	}

	from := strings.TrimSpace(cfg.MailFrom)
	if from == "" {
		from = defaultMailFrom
	}

	sender := &Sender{
		from:           from,
		logoURL:        strings.TrimSpace(cfg.MailLogoURL),
		companyAddress: strings.TrimSpace(cfg.MailCompanyAddress),
		preferencesURL: notificationPreferencesURL(cfg.AppBaseURL),
		templates:      templates,
		send:           sendSMTP,
	}
	if strings.TrimSpace(cfg.MailerDSN) == "" {
		return sender, nil
	}

	smtpCfg, err := parseMailerDSN(cfg.MailerDSN)
	if err != nil {
		return nil, err
	}
	sender.smtp = smtpCfg
	return sender, nil
}

func parseTemplates() (map[string]*template.Template, error) {
	out := make(map[string]*template.Template, len(templateFiles))
	for _, name := range templateFiles {
		tmpl, err := template.ParseFS(
			templateFS,
			"templates/layout.html",
			"templates/"+name+".html",
		)
		if err != nil {
			return nil, fmt.Errorf("parse mail template %s: %w", name, err)
		}
		out[name] = tmpl
	}
	return out, nil
}

func parseMailerDSN(dsn string) (*smtpConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid MAILER_DSN: %w", err)
	}
	if u.Scheme != "smtp" && u.Scheme != "smtps" {
		return nil, fmt.Errorf("invalid MAILER_DSN scheme %q (want smtp)", u.Scheme)
	}
	if u.Scheme == "smtps" {
		return nil, fmt.Errorf("MAILER_DSN smtps is not supported")
	}

	host := u.Host
	if host == "" {
		return nil, fmt.Errorf("invalid MAILER_DSN: missing host")
	}
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), "25")
	}

	cfg := &smtpConfig{addr: host}
	if u.User != nil {
		password, _ := u.User.Password()
		cfg.auth = smtp.PlainAuth("", u.User.Username(), password, u.Hostname())
	}
	return cfg, nil
}

func (s *Sender) Send(ctx context.Context, input port.SendMailInput) error {
	if len(input.To) == 0 {
		return nil
	}
	if s.smtp == nil {
		log.Printf("mail: skipped template=%s to=%d (MAILER_DSN is not configured)", input.TemplateName, len(input.To))
		return nil
	}

	html, err := s.render(input.TemplateName, s.withLayout(input.Subject, input.TemplateData))
	if err != nil {
		return err
	}

	headerFrom, envelopeFrom, err := splitAddress(s.from)
	if err != nil {
		return fmt.Errorf("invalid MAIL_FROM: %w", err)
	}

	msg, err := buildMessage(headerFrom, input.To, input.Subject, html, input.Attachments)
	if err != nil {
		return err
	}
	if err := s.send(ctx, *s.smtp, envelopeFrom, input.To, msg); err != nil {
		return fmt.Errorf("send mail template=%s: %w", input.TemplateName, err)
	}

	log.Printf("mail: sent template=%s to=%d subject=%q via %s", input.TemplateName, len(input.To), input.Subject, s.smtp.addr)
	return nil
}

func notificationPreferencesURL(appBaseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(appBaseURL), "/")
	if base == "" {
		return ""
	}
	return base + "/settings/notifications"
}

func (s *Sender) withLayout(subject string, data map[string]any) map[string]any {
	out := make(map[string]any, len(data)+5)
	for key, value := range data {
		out[key] = value
	}
	out["Subject"] = subject
	if preheader, _ := out["Preheader"].(string); strings.TrimSpace(preheader) == "" {
		out["Preheader"] = subject
	}
	out["LogoURL"] = s.logoURL
	out["CompanyAddress"] = s.companyAddress
	out["PreferencesURL"] = s.preferencesURL
	return out
}

func (s *Sender) render(name string, data map[string]any) (string, error) {
	tmpl, ok := s.templates[name]
	if !ok {
		return "", fmt.Errorf("unknown mail template %q", name)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return "", fmt.Errorf("execute mail template %s: %w", name, err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func splitAddress(value string) (header, envelope string, err error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return "", "", err
	}
	return addr.String(), addr.Address, nil
}

func buildMessage(from string, to []string, subject, html string, attachments []port.MailAttachment) ([]byte, error) {
	if len(attachments) == 0 {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "From: %s\r\n", from)
		fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(to, ", "))
		fmt.Fprintf(&buf, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
		fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
		fmt.Fprintf(&buf, "Content-Type: text/html; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "\r\n")
		buf.WriteString(html)
		return buf.Bytes(), nil
	}

	boundary := "workflow-mail-" + fmt.Sprintf("%d", time.Now().UnixNano())
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=%s\r\n", boundary)
	fmt.Fprintf(&buf, "\r\n")
	fmt.Fprintf(&buf, "--%s\r\n", boundary)
	fmt.Fprintf(&buf, "Content-Type: text/html; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n")
	fmt.Fprintf(&buf, "\r\n")
	buf.WriteString(html)
	fmt.Fprintf(&buf, "\r\n")
	for _, att := range attachments {
		filename := strings.TrimSpace(att.Filename)
		if filename == "" {
			filename = "attachment"
		}
		contentType := strings.TrimSpace(att.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: %s; name=%q\r\n", contentType, filename)
		fmt.Fprintf(&buf, "Content-Disposition: attachment; filename=%q\r\n", filename)
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n")
		fmt.Fprintf(&buf, "\r\n")
		encoded := make([]byte, base64.StdEncoding.EncodedLen(len(att.Bytes)))
		base64.StdEncoding.Encode(encoded, att.Bytes)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			buf.Write(encoded[i:end])
			buf.WriteString("\r\n")
		}
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	return buf.Bytes(), nil
}

func sendSMTP(ctx context.Context, cfg smtpConfig, from string, to []string, msg []byte) error {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", cfg.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	}

	host, _, err := net.SplitHostPort(cfg.addr)
	if err != nil {
		host = cfg.addr
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if cfg.auth != nil {
		if err := client.Auth(cfg.auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

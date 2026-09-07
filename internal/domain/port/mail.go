package port

import "context"

type MailAttachment struct {
	Filename    string
	ContentType string
	Bytes       []byte
}

type SendMailInput struct {
	To           []string
	Subject      string
	TemplateName string
	TemplateData map[string]any
	Attachments  []MailAttachment
}

type MailSender interface {
	Send(ctx context.Context, input SendMailInput) error
}

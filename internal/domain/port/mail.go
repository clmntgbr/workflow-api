package port

import "context"

type SendMailInput struct {
	To           []string
	Subject      string
	TemplateName string
	TemplateData map[string]any
}

type MailSender interface {
	Send(ctx context.Context, input SendMailInput) error
}

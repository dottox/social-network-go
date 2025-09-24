package mailer

import "embed"

const (
	MAX_RETRIES         = 5
	FromName            = "GopherSocial"
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}

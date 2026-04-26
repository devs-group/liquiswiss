package email_adapter

import (
	"crypto/tls"
	"fmt"
	"liquiswiss/config"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/url"

	"github.com/wneessen/go-mail"
)

type smtpAdapter struct {
	cfg      config.Config
	renderer *templateRenderer
}

func newSMTPAdapter(cfg config.Config) IEmailAdapter {
	return &smtpAdapter{cfg: cfg, renderer: defaultRenderer}
}

func (s *smtpAdapter) sendHTML(toAddress, templateName string, content models.EmailContent) error {
	if s.cfg.SMTPHost == "" {
		logger.Logger.Warnw("SMTP host not configured — skipping email send",
			"to", toAddress, "subject", content.Subject)
		return nil
	}

	body, err := s.renderer.render(templateName, content)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	msg := mail.NewMsg()
	if err := msg.FromFormat(s.cfg.SMTPFromName, s.cfg.SMTPFromAddress); err != nil {
		return fmt.Errorf("set From: %w", err)
	}
	if err := msg.To(toAddress); err != nil {
		return fmt.Errorf("set To: %w", err)
	}
	msg.Subject(content.Subject)
	msg.SetBodyString(mail.TypeTextHTML, body)

	clientOpts := []mail.Option{
		mail.WithPort(s.cfg.SMTPPort),
	}

	switch s.cfg.SMTPTLS {
	case "implicit":
		clientOpts = append(clientOpts, mail.WithSSLPort(false), mail.WithTLSPolicy(mail.TLSMandatory))
	case "starttls":
		clientOpts = append(clientOpts, mail.WithTLSPolicy(mail.TLSMandatory))
	default:
		clientOpts = append(clientOpts, mail.WithTLSPolicy(mail.NoTLS))
	}

	if s.cfg.SMTPUser != "" {
		clientOpts = append(clientOpts,
			mail.WithSMTPAuth(mail.SMTPAuthLogin),
			mail.WithUsername(s.cfg.SMTPUser),
			mail.WithPassword(s.cfg.SMTPPass),
		)
	}

	if s.cfg.SMTPTLS != "off" {
		clientOpts = append(clientOpts, mail.WithTLSConfig(&tls.Config{ServerName: s.cfg.SMTPHost}))
	}

	client, err := mail.NewClient(s.cfg.SMTPHost, clientOpts...)
	if err != nil {
		logger.Logger.Errorf("Failed to create SMTP client: %v", err)
		return err
	}

	if err := client.DialAndSend(msg); err != nil {
		logger.Logger.Errorf("Failed to send email: %v", err)
		return err
	}
	return nil
}

func (s *smtpAdapter) SendRegistrationMail(email, code string) error {
	params := url.Values{}
	params.Add("email", email)
	params.Add("code", code)

	content := models.EmailContent{
		Subject:   "Bestätigen Sie Ihre E-Mail",
		PreHeader: "Nur noch ein kleiner Schritt bevor Sie LiquiSwiss nutzen können ...",
		Hello:     "Willkommen bei LiquiSwiss 🇨🇭",
		Content: fmt.Sprintf(
			"Danke für Ihr Interesse an Liquiswiss. Um Ihre Anmeldung abzuschliessen müssen Sie nur noch Ihre E-Mail bestätigen. Bitte beachten Sie, dass dieser Link für maximal %.0f Stunde(n) gültig ist",
			utils.RegistrationCodeValidity.Hours(),
		),
		ButtonText: "E-Mail bestätigen",
		ButtonUrl:  fmt.Sprintf("%s/auth/validate?%s", s.cfg.WebHost, params.Encode()),
		Greetings:  "Wir wünschen Ihnen viel Erfolg<br/>Ihr liquiswiss.ch Team 🚀",
	}
	return s.sendHTML(email, "base.tmpl", content)
}

func (s *smtpAdapter) SendPasswordResetMail(email, code string) error {
	params := url.Values{}
	params.Add("email", email)
	params.Add("code", code)

	content := models.EmailContent{
		Subject:   "Anfrage zum Zurücksetzen des Passworts",
		PreHeader: "",
		Hello:     "Guten Tag! 👋",
		Content: fmt.Sprintf(
			"Sie haben angefordert Ihr Passwort zurückzusetzen. Bitte beachten Sie, dass dieser Link für maximal %.0f Minute(n) gültig ist",
			utils.ResetPasswordDelay.Minutes(),
		),
		ButtonText: "Passwort zurücksetzen",
		ButtonUrl:  fmt.Sprintf("%s/auth/reset-password?%s", s.cfg.WebHost, params.Encode()),
		Greetings:  "Sollten Sie dies nicht beantragt haben, können Sie diese E-Mail ignorieren.<br/><br/>Wir wünschen Ihnen weiterhin viel Erfolg<br/>Ihr liquiswiss.ch Team 🚀",
	}
	return s.sendHTML(email, "base.tmpl", content)
}

func (s *smtpAdapter) SendInvitationMail(email, token, organisationName, invitedByName string) error {
	params := url.Values{}
	params.Add("token", token)

	content := models.EmailContent{
		Subject:   fmt.Sprintf("Einladung zu %s auf LiquiSwiss", organisationName),
		PreHeader: fmt.Sprintf("%s hat Sie eingeladen ...", invitedByName),
		Hello:     "Guten Tag! 👋",
		Content: fmt.Sprintf(
			"%s hat Sie eingeladen, der Organisation <strong>%s</strong> auf LiquiSwiss beizutreten. Klicken Sie auf den Button unten, um die Einladung anzunehmen. Bitte beachten Sie, dass dieser Link für maximal %.0f Tag(e) gültig ist.",
			invitedByName,
			organisationName,
			utils.InvitationValidity.Hours()/24,
		),
		ButtonText: "Einladung annehmen",
		ButtonUrl:  fmt.Sprintf("%s/auth/invitation?%s", s.cfg.WebHost, params.Encode()),
		Greetings:  "Wir wünschen Ihnen viel Erfolg<br/>Ihr liquiswiss.ch Team 🚀",
	}
	return s.sendHTML(email, "base.tmpl", content)
}

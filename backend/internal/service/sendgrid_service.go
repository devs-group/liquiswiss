//go:generate mockgen -package=mocks -destination ../mocks/sendgrid_service.go liquiswiss/internal/service ISendgridService
package service

import (
	"errors"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"liquiswiss/config"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"net/url"
)

type SendgridService struct {
	apiKey string
}

type ISendgridService interface {
	SendMail(from *mail.Email, to *mail.Email, templateId string, dynamicTemplateData interface{}) error
	SendRegistrationMail(email, code string) error
	SendPasswordResetMail(email, code string) error
}

func NewSendgridService(apiKey string) ISendgridService {
	return &SendgridService{
		apiKey: apiKey,
	}
}

func (s SendgridService) SendMail(from *mail.Email, to *mail.Email, templateId string, templateData interface{}) error {
	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.SetTemplateID(templateId)

	p := mail.NewPersonalization()
	p.AddTos(to)

	templateMap, err := utils.StructToMap(templateData)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	p.DynamicTemplateData = templateMap

	m.AddPersonalizations(p)

	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(m)
	if err != nil {
		logger.Logger.Error(err)
		return err
	} else {
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			logger.Logger.Error(response.StatusCode, response.Body)
			return errors.New(response.Body)
		}
	}
	return nil
}

func (s SendgridService) SendRegistrationMail(email, code string) error {
	cfg := config.GetConfig()

	params := url.Values{}
	params.Add("email", email)
	params.Add("code", code)

	err := s.SendMail(
		&mail.Email{
			Name:    "LiquiSwiss",
			Address: "no-reply@liquiswiss.ch",
		},
		&mail.Email{
			Name:    "",
			Address: email,
		},
		utils.RegistrationMailTemplate,
		models.RegistrationMail{
			Subject:   "Bestätigen Sie Ihre E-Mail",
			PreHeader: "Nur noch ein kleiner Schritt bevor Sie LiquiSwiss nutzen können ...",
			Hello:     "Willkommen bei LiquiSwiss 🇨🇭",
			Content: fmt.Sprintf(
				"Danke für Ihr Interesse an Liquiswiss. Um Ihre Anmeldung abzuschliessen müssen Sie nur noch Ihre E-Mail bestätigen. Bitte beachten Sie, dass dieser Link für maximal %.0f Stunde(n) gültig ist",
				utils.RegistrationCodeValidity.Hours(),
			),
			ButtonText: "E-Mail bestätigen",
			ButtonUrl:  fmt.Sprintf("%s/auth/validate?%s", cfg.WebHost, params.Encode()),
			Greetings:  "Wir wünschen Ihnen viel Erfolg<br/>Ihr liquiswiss.ch Team 🚀",
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (s SendgridService) SendPasswordResetMail(email, code string) error {
	cfg := config.GetConfig()

	params := url.Values{}
	params.Add("email", email)
	params.Add("code", code)

	err := s.SendMail(
		&mail.Email{
			Name:    "LiquiSwiss",
			Address: "no-reply@liquiswiss.ch",
		},
		&mail.Email{
			Name:    "",
			Address: email,
		},
		utils.RegistrationMailTemplate,
		models.RegistrationMail{
			Subject:   "Anfrage zum Zurücksetzen des Passworts",
			PreHeader: "",
			Hello:     "Guten Tag! 👋",
			Content: fmt.Sprintf(
				"Sie haben angefordert Ihr Passwort zurückzusetzen. Bitte beachten Sie, dass dieser Link für maximal %.0f Minute(n) gültig ist",
				utils.ResetPasswordDelay.Minutes(),
			),
			ButtonText: "Passwort zurücksetzen",
			ButtonUrl:  fmt.Sprintf("%s/auth/reset-password?%s", cfg.WebHost, params.Encode()),
			Greetings:  "Sollten Sie dies nicht beantragt haben, können Sie diese E-Mail ignorieren.<br/><br/>Wir wünschen Ihnen weiterhin viel Erfolg<br/>Ihr liquiswiss.ch Team 🚀",
		},
	)
	if err != nil {
		return err
	}
	return nil
}

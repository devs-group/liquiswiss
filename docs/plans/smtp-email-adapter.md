# SMTP Email Adapter

## Goal

Replace SendGrid API dependency with standard SMTP for local development. Use Mailpit as local mail server. Keep SendGrid as an option for production while sharing a common interface.

## Motivation

- Remove external API dependency for local development
- Open source friendly - no API keys required to run locally
- Mailpit provides a web UI to view sent emails during development
- Flexibility to switch between email providers

## Current State

- `ISendgridAdapter` interface in `backend/internal/adapter/sendgrid_adapter/sendgrid_adapter.go`
- Uses SendGrid API with dynamic templates
- Email content defined in `models.SendgridMail` struct
- Three email types: Registration, Password Reset, Invitation

## Architecture

### New Interface: `IEmailAdapter`

Create a provider-agnostic email interface:

```go
// backend/internal/adapter/email/email.go
package email

type EmailContent struct {
    Subject    string
    PreHeader  string
    Hello      string
    Content    string
    ButtonText string
    ButtonUrl  string
    Greetings  string
}

type Email struct {
    FromName    string
    FromAddress string
    ToName      string
    ToAddress   string
    Content     EmailContent
}

type IEmailAdapter interface {
    Send(email Email) error
    SendRegistrationMail(email, code string) error
    SendPasswordResetMail(email, code string) error
    SendInvitationMail(email, token, organisationName, invitedByName string) error
}
```

### Implementations

| Adapter | Package | Use Case |
|---------|---------|----------|
| SMTP | `email/smtp` | Local dev (Mailpit), self-hosted SMTP |
| SendGrid | `email/sendgrid` | Production (existing) |

### HTML Email Template

Create a single responsive HTML template with 90%+ email client compatibility:

```
backend/internal/adapter/email/templates/
  base.html       # Main template with placeholders
```

**Template Requirements:**
- Table-based layout (Outlook compatibility)
- Inline CSS only (Gmail compatibility)
- Max width 600px
- System fonts with fallbacks
- Tested against: Gmail, Outlook, Apple Mail, Yahoo

**Placeholders:**
- `{{.Subject}}`
- `{{.PreHeader}}`
- `{{.Hello}}`
- `{{.Content}}`
- `{{.ButtonText}}`
- `{{.ButtonUrl}}`
- `{{.Greetings}}`

### SMTP Adapter Implementation

```go
// backend/internal/adapter/email/smtp/smtp.go
package smtp

type SMTPConfig struct {
    Host     string
    Port     int
    Username string // Optional for Mailpit
    Password string // Optional for Mailpit
    From     string
}

type SMTPAdapter struct {
    config   SMTPConfig
    template *template.Template
}

func NewSMTPAdapter(config SMTPConfig) (email.IEmailAdapter, error) {
    // Load and parse HTML template
    // Return adapter
}

func (s *SMTPAdapter) Send(email email.Email) error {
    // Render HTML from template
    // Send via SMTP using net/smtp or go-mail
}
```

### Configuration

Add to `backend/config/config.go`:

```go
type Config struct {
    // Existing...

    // Email provider: "smtp" or "sendgrid"
    EmailProvider string `env:"EMAIL_PROVIDER" envDefault:"smtp"`

    // SMTP settings (for EMAIL_PROVIDER=smtp)
    SMTPHost     string `env:"SMTP_HOST" envDefault:"localhost"`
    SMTPPort     int    `env:"SMTP_PORT" envDefault:"1025"`
    SMTPUsername string `env:"SMTP_USERNAME"`
    SMTPPassword string `env:"SMTP_PASSWORD"`
    SMTPFrom     string `env:"SMTP_FROM" envDefault:"no-reply@liquiswiss.ch"`
}
```

### Factory Function

```go
// backend/internal/adapter/email/factory.go
func NewEmailAdapter(cfg *config.Config) (IEmailAdapter, error) {
    switch cfg.EmailProvider {
    case "sendgrid":
        return sendgrid.NewSendgridAdapter(cfg.SendgridApiKey, cfg.SendgridTemplateID)
    case "smtp":
        return smtp.NewSMTPAdapter(smtp.SMTPConfig{
            Host:     cfg.SMTPHost,
            Port:     cfg.SMTPPort,
            Username: cfg.SMTPUsername,
            Password: cfg.SMTPPassword,
            From:     cfg.SMTPFrom,
        })
    default:
        return nil, fmt.Errorf("unknown email provider: %s", cfg.EmailProvider)
    }
}
```

## Docker Compose

Add Mailpit service:

```yaml
# docker-compose.yml
services:
  mailpit:
    image: axllent/mailpit
    container_name: mailpit
    restart: unless-stopped
    ports:
      - "8025:8025"  # Web UI
      - "1025:1025"  # SMTP
```

## Files to Create

| File | Purpose |
|------|---------|
| `backend/internal/adapter/email/email.go` | Common interface and types |
| `backend/internal/adapter/email/factory.go` | Provider factory |
| `backend/internal/adapter/email/smtp/smtp.go` | SMTP implementation |
| `backend/internal/adapter/email/sendgrid/sendgrid.go` | Refactored SendGrid |
| `backend/internal/adapter/email/templates/base.html` | HTML email template |

## Files to Modify

| File | Changes |
|------|---------|
| `backend/config/config.go` | Add SMTP and provider config |
| `backend/main.go` | Use factory to create email adapter |
| `backend/.env.example` | Add SMTP env vars |
| `docker-compose.yml` | Add Mailpit service |
| `backend/pkg/models/mail.go` | Rename `SendgridMail` to `EmailContent` (or keep for compatibility) |

## Migration Steps

1. Create new `email` package with interface
2. Create SMTP adapter with HTML template
3. Refactor SendGrid adapter to implement new interface
4. Add factory function
5. Update `main.go` to use factory
6. Add Mailpit to docker-compose
7. Update documentation
8. Test both providers

## HTML Template Design

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Subject}}</title>
  <!--[if mso]>
  <style type="text/css">
    table { border-collapse: collapse; }
    .button { padding: 12px 24px !important; }
  </style>
  <![endif]-->
</head>
<body style="margin:0; padding:0; background-color:#f4f4f4; font-family:Arial,Helvetica,sans-serif;">
  <!-- Preheader -->
  <div style="display:none; max-height:0; overflow:hidden;">{{.PreHeader}}</div>

  <!-- Main table -->
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background-color:#f4f4f4;">
    <tr>
      <td align="center" style="padding:20px 0;">
        <table role="presentation" width="600" cellspacing="0" cellpadding="0" style="background-color:#ffffff; border-radius:8px;">
          <!-- Logo -->
          <tr>
            <td align="center" style="padding:30px 40px 20px;">
              <h1 style="margin:0; color:#1a1a1a; font-size:24px;">LiquiSwiss</h1>
            </td>
          </tr>
          <!-- Hello -->
          <tr>
            <td style="padding:0 40px 10px;">
              <h2 style="margin:0; color:#1a1a1a; font-size:20px;">{{.Hello}}</h2>
            </td>
          </tr>
          <!-- Content -->
          <tr>
            <td style="padding:0 40px 20px; color:#555555; font-size:16px; line-height:24px;">
              {{.Content}}
            </td>
          </tr>
          <!-- Button -->
          <tr>
            <td align="center" style="padding:10px 40px 30px;">
              <table role="presentation" cellspacing="0" cellpadding="0">
                <tr>
                  <td style="background-color:#0066cc; border-radius:4px;">
                    <a href="{{.ButtonUrl}}" target="_blank" style="display:inline-block; padding:14px 28px; color:#ffffff; text-decoration:none; font-size:16px; font-weight:bold;">{{.ButtonText}}</a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <!-- Greetings -->
          <tr>
            <td style="padding:0 40px 30px; color:#555555; font-size:16px; line-height:24px;">
              {{.Greetings}}
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
```

## Testing

- Unit tests for SMTP adapter (mock SMTP server)
- Integration test with Mailpit
- Visual testing of HTML template across email clients (use Litmus or Email on Acid, or manual testing)

## Environment Examples

### Local Development (.env)
```
EMAIL_PROVIDER=smtp
SMTP_HOST=localhost
SMTP_PORT=1025
```

### Production (.env)
```
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=SG.xxx
SENDGRID_TEMPLATE_ID=d-xxx
```

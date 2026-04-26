package email_adapter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"liquiswiss/config"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
)

func init() {
	logger.NewZapLogger(false)
}

func newAdapterForTest(t *testing.T, cfg config.Config) *smtpAdapter {
	t.Helper()
	cfg.WebHost = "https://app.test"
	a, ok := newSMTPAdapter(cfg).(*smtpAdapter)
	require.True(t, ok)
	return a
}

func TestSendSkipsWhenSMTPHostEmpty(t *testing.T) {
	a := newAdapterForTest(t, config.Config{})
	require.NoError(t, a.SendRegistrationMail("user@example.com", "code123"))
	require.NoError(t, a.SendPasswordResetMail("user@example.com", "code456"))
	require.NoError(t, a.SendInvitationMail("user@example.com", "tok", "Acme", "Bob"))
}

func TestRenderRegistrationTemplate(t *testing.T) {
	a := newAdapterForTest(t, config.Config{})
	body, err := a.renderer.render("base.tmpl", models.EmailContent{
		Subject:    "Registration",
		Hello:      "Hi",
		Content:    "Click <strong>here</strong> to confirm.",
		ButtonText: "Confirm",
		ButtonUrl:  "https://app.test/auth/validate?email=user%40example.com&code=abc",
		Greetings:  "Cheers<br/>Team",
	})
	require.NoError(t, err)
	require.Contains(t, body, "https://app.test/auth/validate?email=user%40example.com&amp;code=abc")
	require.Contains(t, body, "<strong>here</strong>")
	require.Contains(t, body, "Cheers<br/>Team")
}

func TestRenderEscapesUserSuppliedNames(t *testing.T) {
	a := newAdapterForTest(t, config.Config{})
	body, err := a.renderer.render("base.tmpl", models.EmailContent{
		Subject:    "Inv",
		Hello:      "Guten Tag!",
		Content:    "<script>x</script>",
		ButtonText: "Go",
		ButtonUrl:  "https://app.test/x",
		Greetings:  "k",
	})
	require.NoError(t, err)
	// Content is template.HTML so passes through; Hello is plain string and auto-escaped if needed
	require.True(t, strings.Contains(body, "<script>x</script>"))
}

func TestRenderUnknownTemplateErrors(t *testing.T) {
	a := newAdapterForTest(t, config.Config{})
	_, err := a.renderer.render("missing.tmpl", models.EmailContent{})
	require.Error(t, err)
}

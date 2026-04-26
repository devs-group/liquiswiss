package email_adapter

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"liquiswiss/pkg/models"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type templateRenderer struct {
	tmpl *template.Template
}

var defaultRenderer = mustLoadTemplates()

func mustLoadTemplates() *templateRenderer {
	tmpl, err := template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		panic(fmt.Errorf("parse email templates: %w", err))
	}
	return &templateRenderer{tmpl: tmpl}
}

type templateData struct {
	Subject    string
	PreHeader  string
	Hello      string
	Content    template.HTML
	ButtonText string
	ButtonUrl  string
	Greetings  template.HTML
}

func (r *templateRenderer) render(name string, content models.EmailContent) (string, error) {
	t := r.tmpl.Lookup(name)
	if t == nil {
		return "", fmt.Errorf("template %q not found", name)
	}
	data := templateData{
		Subject:    content.Subject,
		PreHeader:  content.PreHeader,
		Hello:      content.Hello,
		Content:    template.HTML(content.Content),
		ButtonText: content.ButtonText,
		ButtonUrl:  content.ButtonUrl,
		Greetings:  template.HTML(content.Greetings),
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template %q: %w", name, err)
	}
	return buf.String(), nil
}

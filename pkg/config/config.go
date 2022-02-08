package config

import (
	"html/template"

	"github.com/alexedwards/scs/v2"
)

type TemplateCache map[string]*template.Template

type AppConfig struct {
	UseCache      bool
	TemplateCache TemplateCache
	InProduction  bool
	Session       *scs.SessionManager
}

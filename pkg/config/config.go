package config

import (
	"github.com/amiranbari/bookings/pkg/models"
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

type TemplateCache map[string]*template.Template

type AppConfig struct {
	UseCache      bool
	TemplateCache TemplateCache
	InProduction  bool
	Session       *scs.SessionManager
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	MailChan      chan models.MailData
}

package models

import "github.com/jofosuware/small-business-management-app/internal/forms"

type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]string
	FloatMap        map[string]string
	Data            map[string]interface{}
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	Form            *forms.Form
	IsAuthenticated int
}

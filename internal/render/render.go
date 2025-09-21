package render

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/helpers"
	"github.com/jofosuware/small-business-management-app/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig
var pathToTemplates = "./templates"

var functions = template.FuncMap{
	"humanDate":       HumanDate,
	"formatDate":      FormatDate,
	"convertToBase64": ConvertToBase64,
	"toDecimalPlace":  helpers.ToDecimalPlace,
}

// NewRenderer sets the config for the templates package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// ConvertToBase64 encodes bytes to base64 string
func ConvertToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// HumanDate returns time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("02-01-2006 3:04 pm")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
		td.Data["user"] = app.Session.Get(r.Context(), "user").(models.User)
	}
	return td
}

// Template renders templates using html/templates
func Template(w http.ResponseWriter, r *http.Request, html string, td *models.TemplateData) error {
	var tc map[string]*template.Template
	var err error

	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		// this is just used for testing, so that we rebuild
		// the cache on every request
		tc, err = CreateTemplateCache()
		if err != nil {
			log.Println("error creating template cache", err)
			return err
		}
	}

	t, ok := tc[html]
	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	err = t.Execute(buf, td)
	if err != nil {
		log.Printf("error executing template: %v", err)
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}

	return nil
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Get all the files ending with *.page.html
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Get all the files ending with *.layout.html
	layouts, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// Range through all the pages
	for _, page := range pages {
		name := filepath.Base(page)

		// Create a new template set with the page's name and parse the page file
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// If layouts exist, parse them into the template set
		if len(layouts) > 0 {
			_, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
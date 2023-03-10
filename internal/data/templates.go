package data

import (
	"html/template"
	"path/filepath"
	"time"
)

type TemplateData struct {
	// Form            *forms.Form
	// Snippet         *models.Snippet
	IsAuthenticated bool
	Envelope        Envelope
	CurrentYear     string
	ErrorText       string
	Code            int
	User            User
	Tickets         []Ticket
}

type Envelope map[string]interface{}

// Initialize a template.FuncMap object and store it in a global variable. This is essentially
// a string-keyed map which acts as a lookup between the names of our custom template
// functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func NewTemplateCache(dir string) (map[string]*template.Template, error) {
	// init cache
	cache := map[string]*template.Template{}

	// Use the filepath.Glob function to get a slice of all file
	// paths with the extension '.page.gohtml'. This essentially gives
	// use a slice of all the 'page' templates for the application.
	var pages []string
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		return nil, err
	}

	// Loop through the pages one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.page.gohtml') from the full
		// file path and assign it to the name variable.
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before you
		// call the ParseFiles() method. This means we have to use template.New() to
		// create an empty template set, use the Funcs() method to register the
		// template.FuncMap, and then parse the file as normal.
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method to add any 'layout' templates to the template set
		// (in our case, it's just the 'base' layout currently).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.html"))
		if err != nil {
			return nil, err
		}

		// Use the ParseGlob method again to add any 'partial' templates to the template
		// set (in our case, it's just the 'footer' partial currently).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.html"))
		if err != nil {
			return nil, err
		}

		// set file name to it's template
		// cache["home.page.html"] = template
		cache[name] = ts
	}

	return cache, nil
}

// humanDate returns a nicely formatted human-readable string representation of time.Time.
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// Convert the time to UTC before formatting it.
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

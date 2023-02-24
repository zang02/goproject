package main

import (
	"app/internal/data"
	"app/internal/validator"
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *data.TemplateData) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.gohtml'). If no entry exists in the cache with the provided name,
	// call the serverError helper method.
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	// Initialize a new buffer.
	buff := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the http.ResponseWriter.
	// If there is an error, call our serverError helper and then return.
	td.CurrentYear = fmt.Sprintf("%v", time.Now().Year())

	if td.Code == 0 {
		td.Code = 200
	}
	// td.Flash = app.session.PopString(r, "flash")
	// td.IsAuthenticated = app.isAuthenticated(r)
	// td.CSRFToken = nosurf.Token(r)

	err := ts.Execute(buff, td)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write the contents of the buffer to the http.ResponseWriter. Again, this is another place
	// where we pass our http.ResponseWriter to a function that take an io.Writer
	if _, err = buff.WriteTo(w); err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.PrintError(err.Error(), "server error")

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(len(email) < 5000, "email", "must not be more than 5000 bytes long")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}
func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
func ValidateUser(v *validator.Validator, user *data.User) {
	v.Check(user.Login != "", "login", "must be provided")
	v.Check(len(user.Login) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(user.Name != "", "name", "must be provided")
	ValidateEmail(v, user.Email)
	ValidatePassword(v, user.Password)
	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the user). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
}

func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		//first thing to do when working with goroutines  is to defer the wg.Done())))))))))))
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err).Error(), "")
			}
		}()

		fn()
	}()
}

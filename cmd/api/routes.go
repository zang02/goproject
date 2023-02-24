package main

import (
	"app/internal/data"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.requireAuth)

	r := mux.NewRouter()
	templateData := &data.TemplateData{}

	r.Handle("/", app.authenticate(templateData, app.Render("home.page.html", templateData))).Methods("GET")
	r.Handle("/testCookie", app.testCookie())

	r.Handle("/signup", app.authenticate(templateData, app.Render("signup.page.html", templateData))).Methods("GET")
	r.Handle("/login", app.authenticate(templateData, app.Render("login.page.html", templateData))).Methods("GET")
	r.Handle("/signup", app.authenticate(templateData, app.signupHandler())).Methods("POST")
	r.Handle("/login", app.authenticate(templateData, app.loginHandler())).Methods("POST")

	r.Handle("/logout", dynamicMiddleware.ThenFunc(app.logoutHandler)).Methods("POST")

	r.Handle("/receipt/:id", app.showTicketHandler())
	r.Handle("/receipt", app.GetAllTickets())

	r.Handle("/product", app.GroceryStorehandle())

	// gorilla mux file server
	fileServer := http.FileServer(http.Dir("./ui/static"))
	r.PathPrefix("/").Handler(http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(r)
}

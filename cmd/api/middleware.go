package main

import (
	"app/internal/data"
	"fmt"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.PrintInfo(fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI()), "")
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "http://localhost:"+app.config.port, http.StatusSeeOther)
			return
		}
		_, err = app.models.Tokens.GetTokenDocumentByToken(tokenCookie.Value)
		if err != nil {
			app.logoutHandler(w, r)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(td *data.TemplateData, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		td.User.Email = ""
		td.User.Login = ""
		td.User.Name = ""
		td.User.CreateDate = ""
		td.IsAuthenticated = false
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		TokenDocument, err := app.models.Tokens.GetTokenDocumentByToken(tokenCookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := app.models.Users.GetByLogin(TokenDocument.UserLogin)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		} else {
			td.User.Email = user.Email
			td.User.Login = user.Login
			td.User.Name = user.Name
			td.User.CreateDate = user.CreateDate
			td.IsAuthenticated = true
		}

		next.ServeHTTP(w, r)
	})
}

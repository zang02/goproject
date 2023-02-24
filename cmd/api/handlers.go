package main

import (
	"app/internal/data"
	"app/internal/validator"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) testCookie() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie2 := &http.Cookie{
			Name:     "page",
			Value:    "GoLinuxCloud",
			Expires:  time.Now().Add(365 * 24 * time.Hour),
			Secure:   false,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, cookie2)

		fmt.Fprintln(w, r.Cookies())
	})
}

// enter requireAuth middleware
// pointer template data set to zero value
// then get user by token
// add user data to pointer template data

func (app *application) Render(templateName string, templateData *data.TemplateData) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, templateName, templateData)
	})
}

func (app *application) signupHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		hashedPw, _ := bcrypt.GenerateFromPassword([]byte(r.Form["password"][0]), 12)

		user := data.User{
			Email:    r.Form["email"][0],
			Login:    r.Form["login"][0],
			Password: string(hashedPw),
			Name:     r.Form["name"][0],
		}
		v := validator.New()
		ValidateUser(v, &user)
		if !v.Valid() {
			errMsg := ""
			for k, v := range v.Errors {
				errMsg += k + " " + v + "\n"
			}
			app.render(w, r, "signup.page.html", &data.TemplateData{
				ErrorText: errMsg,
				Code:      409,
			})
			return
		}

		err := app.models.Users.Insert(user)
		if err != nil {
			if strings.Contains(err.Error(), "login") {
				app.render(w, r, "signup.page.html", &data.TemplateData{
					ErrorText: "user with such login already exists",
					Code:      409,
				})
				return
			}
			if strings.Contains(err.Error(), "email") {
				app.render(w, r, "signup.page.html", &data.TemplateData{
					ErrorText: "email already in use",
					Code:      409,
				})
				return
			}

		}

		cookie := http.Cookie{
			Name:     "token",
			Value:    user.Login,
			Expires:  time.Now().Add(365 * 24 * time.Hour),
			Secure:   true,
			HttpOnly: true,
			Path:     "/",
			SameSite: 4,
		}

		http.SetCookie(w, &cookie)
		app.background(func() {
			app.models.Tokens.Insert(data.Token{
				Token:     cookie.Value,
				UserLogin: user.Login,
			})
		})
		http.Redirect(w, r, "http://localhost:"+app.config.port, http.StatusCreated)
	})
}

func (app *application) loginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		login := r.Form["login"][0]
		password := r.Form["password"][0]

		if login == "" || password == "" {
			app.render(w, r, "login.page.html", &data.TemplateData{
				ErrorText: "login and/or password empty",
				Code:      409,
			})
			return
		}
		user, err := app.models.Users.GetByLogin(login)
		if err == errors.New("mongo: no documents in result") {
			app.render(w, r, "login.page.html", &data.TemplateData{
				ErrorText: "user not found",
				Code:      400,
			})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			app.render(w, r, "login.page.html", &data.TemplateData{
				ErrorText: "wrong password",
				Code:      409,
			})
			return
		}
		cookie := &http.Cookie{
			Name:     "token",
			Value:    user.Login,
			Expires:  time.Now().Add(365 * 24 * time.Hour),
			Secure:   true,
			HttpOnly: true,
			Path:     "/",
			SameSite: 4,
		}

		http.SetCookie(w, cookie)
		app.background(func() {
			app.models.Tokens.Insert(data.Token{
				Token:     cookie.Value,
				UserLogin: user.Login,
			})
		})

		http.Redirect(w, r, "http://localhost:"+app.config.port, http.StatusOK)

	})
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "http://localhost:"+app.config.port, http.StatusSeeOther)
		return
	}

	app.background(func() {
		app.models.Tokens.DeleteToken(cookie.Value)
	})
	cookie.MaxAge = -1
	cookie.Expires = time.Now()
	cookie.Value = ""
	http.SetCookie(w, cookie)
	// app.background(
	// 	func()func(){
	// 		app.models.Tokens.GetTokenDocumentByToken(cookie.Value)
	// 		app.models.Tokens.Drop()
	// 		return
	// 	}(),
	// )
	// claimsMap := map[string]string{
	// 	"aud": app.config.port,
	// 	"iss": user.Login,
	// 	"exp": fmt.Sprint(time.Now().Add(time.Minute * 1).Unix()),
	// }
	// secret := "secret"
	// header := "HS256"
	// tokenString, err := jwt.GenerateToken(header, claimsMap, secret)

	http.Redirect(w, r, "http://localhost:"+app.config.port, http.StatusOK)
}
func (app *application) createTicketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		price, _ := strconv.Atoi(r.Form["price"][0])
		amount, _ := strconv.Atoi(r.Form["amount"][0])

		app.models.Tickets.Insert(data.Ticket{
			UserLogin: r.Form["login"][0],
			Products: []data.Product{
				data.Product{
					Name:   r.Form["productName"][0],
					Price:  price,
					Amount: amount,
				},
			},
		})
	})
}

func (app *application) showTicketHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ticketID := vars["id"]

		ticket, err := app.models.Tickets.GetById(ticketID)
		if err != nil {
			fmt.Println(err)
			app.render(w, r, "tickets.page.html", &data.TemplateData{
				ErrorText: "No tickets yet",
			})
			return
		}
		app.render(w, r, "tickets.page.html", &data.TemplateData{
			Tickets: []data.Ticket{
				ticket,
			},
		})
	})
}
func (app *application) GetAllTickets() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tickets, err := app.models.Tickets.GetLatest()
		if err != nil {
			fmt.Println(err)
			app.render(w, r, "tickets.page.html", &data.TemplateData{
				ErrorText: "No tickets yet",
			})
			return
		}
		app.render(w, r, "tickets.page.html", &data.TemplateData{
			Tickets: tickets,
		})
	})
}

func (app *application) GroceryStorehandle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, "ticketCreate.page.html", &data.TemplateData{})
	})
}

// GetByLogin(string)
// GetAllUsers() ([]User, error)
// DeleteUserByLogin(login string)
// UpdateUserByLogin(login string, newUser User)

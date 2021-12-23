package session

import (
	"net/http"
)

func (d Device) Authenticate() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Bind()
		m.Authenticate()

		NewView(w, d.domain).
			SetUser(m).
			SetCookie(m, d.cookieName).
			Error(m)
	})
}

func (d Device) CreateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.CreateUser()
		NewView(w, d.domain).Error(m)
	})
}

func (d Device) Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Signin()

		NewView(w, d.domain).
			SetUser(m).
			JSON(m)
	})
}

func (d Device) Signout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.BindUser()
		m.Signout()

		NewView(w, d.domain).
			SetUser(m).
			Error(m)
	})
}

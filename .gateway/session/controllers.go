package session

import (
	"net/http"
)

func (d Device) Authenticate() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Bind()
		m.Authenticate()

		v := NewView(w)
		v.SetUser(m)
		v.SetCookie(m)
	})
}

func (d Device) CreateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.CreateUser()
		NewView(w).Error(m)
	})
}

func (d Device) Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Signin()

		v := NewView(w)
		v.SetUser(m)
		v.JSON(m)
	})
}

func (d Device) Signout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.BindUser()
		m.Signout()

		v := NewView(w)
		v.SetUser(m)
	})
}

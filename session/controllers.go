package session

import (
	"fmt"
	"net/http"
)

func (d Device) Authenticate() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Bind()
		m.Authenticate()

		NewView(w).
			SetUser(m).
			SetCookie(m).
			Error(m)
	})
}

func (d Device) CreateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("return func")
		m := NewModel(r, d.Store)
		m.CreateUser()
		NewView(w).Error(m)
	})
}

func (d Device) Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Signin()

		NewView(w).
			SetUser(m).
			JSON(m)
	})
}

func (d Device) Signout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.BindUser()
		m.Signout()

		NewView(w).
			SetUser(m).
			Error(m)
	})
}

package session

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/oligoden/chassis/device/view"
)

type View struct {
	view.Default
	secure bool
}

func NewView(w http.ResponseWriter) *View {
	v := &View{}
	v.Default = view.Default{}
	v.Response = w
	if os.Getenv("SECURE") == "true" {
		v.secure = true
	}
	return v
}

func (v View) SetUser(m *Model) {
	if m.Err() != nil {
		log.Println("error setting X_user", m.Err())
		return
	}

	fmt.Println("setting response X_user", m.user)
	v.Response.Header().Set("X_user", m.user)
}

func (v View) SetCookie(m *Model) {
	if m.Err() != nil {
		return
	}

	expire := time.Now().Add(24 * 200 * time.Hour)
	cookie := &http.Cookie{
		Name:     "session",
		Value:    m.session,
		Path:     "/",
		Expires:  expire,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   v.secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(v.Response, cookie)
}

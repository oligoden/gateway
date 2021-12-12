package session

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/oligoden/chassis/device/view"
)

type View struct {
	view.Default
	secure bool
	domain string
}

func NewView(w http.ResponseWriter, domain string) *View {
	v := &View{}
	v.Default = view.Default{}
	v.Response = w
	if os.Getenv("SECURE") == "true" {
		v.secure = true
	}
	v.domain = domain
	return v
}

func (v *View) SetUser(m *Model) *View {
	if m.Err() != nil {
		return v
	}

	fmt.Println("setting response X_user", m.user)
	v.Response.Header().Set("X_user", m.user)
	return v
}

func (v *View) SetCookie(m *Model) *View {
	if m.Err() != nil {
		return v
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
		Domain:   v.domain,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(v.Response, cookie)
	return v
}

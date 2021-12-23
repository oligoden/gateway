package session

import (
	"net/http"

	"github.com/oligoden/chassis/device"
	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/view"
)

type Device struct {
	device.Default
	domain     string
	cookieName string
}

func NewDevice(s model.Connector, domain string) *Device {
	d := Device{
		domain:     domain,
		cookieName: "session",
	}
	nm := func(r *http.Request) model.Operator { return NewModel(r, s) }
	nv := func(w http.ResponseWriter) view.Operator { return NewView(w, domain) }
	d.Default = device.NewDevice(nm, nv, s)
	return &d
}

func (d *Device) SetCookieName(n string) {
	d.cookieName = n
}

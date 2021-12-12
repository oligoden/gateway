package session

import (
	"net/http"

	"github.com/oligoden/chassis/device"
	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/view"
)

type Device struct {
	device.Default
	domain string
}

func NewDevice(s model.Connector, domain string) *Device {
	d := &Device{}
	nm := func(r *http.Request) model.Operator { return NewModel(r, s) }
	nv := func(w http.ResponseWriter) view.Operator { return NewView(w, domain) }
	d.Default = device.NewDevice(nm, nv, s)
	return d
}

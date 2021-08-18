package session

import (
	"net/http"

	"github.com/oligoden/chassis/device"
	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/view"
)

type Device struct {
	device.Default
}

func NewDevice(s model.Connector) *Device {
	d := &Device{}
	nm := func(r *http.Request) model.Operator { return NewModel(r, s) }
	nv := func(w http.ResponseWriter) view.Operator { return NewView(w) }
	d.Default = device.NewDevice(nm, nv, s)
	return d
}

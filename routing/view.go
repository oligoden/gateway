package routing

import (
	"net/http"

	"github.com/oligoden/chassis/device/view"
)

type View struct {
	view.Default
}

func NewView(w http.ResponseWriter) *View {
	v := &View{}
	v.Default = view.Default{}
	v.Response = w
	return v
}

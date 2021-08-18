package subdomain

import (
	"net/http"

	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/model/data"
)

type Model struct {
	model.Default
}

func NewModel(r *http.Request, s model.Connector) *Model {
	m := &Model{}
	m.Default = model.Default{}
	m.Request = r
	m.Store = s
	m.BindUser()
	m.NewData = func() data.Operator { return NewRecord() }
	m.Data(NewRecord())
	return m
}

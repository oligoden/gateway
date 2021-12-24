package routing

import (
	"net/http"

	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/model/data"
	"github.com/oligoden/chassis/storage/gosql"
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

func (m *Model) URL() string {
	if m.Err() != nil {
		return ""
	}

	e := NewRecord()
	c := m.Store.Connect(m.User())
	whereDomain := gosql.NewWhere("domain=?", m.Request.Host)
	// wherePath := gosql.NewWhere("path=?", m.Request.URL.Path)
	c.AddModifiers(whereDomain)
	c.Read(e)
	if c.Err() != nil {
		m.Err(c.Err())
		return ""
	}

	return e.URL
}

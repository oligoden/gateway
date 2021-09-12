package subdomain

import (
	"net/http"
	"strings"

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

func (m *Model) URL(dmn string) string {
	if m.Err() != nil {
		return ""
	}

	subdomain := strings.TrimSuffix(m.Request.Host, "."+dmn)
	e := NewRecord()
	c := m.Store.Connect(m.User())
	where := gosql.NewWhere("subdomain=?", subdomain)
	c.AddModifiers(where)
	c.Read(e)
	if c.Err() != nil {
		m.Err(c.Err())
		return ""
	}

	return e.URL
}

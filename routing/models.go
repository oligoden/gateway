package routing

import (
	"net/http"
	"sort"
	"strings"

	"github.com/oligoden/chassis/device/model"
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
	m.Data(NewRecord())
	return m
}

func (m *Model) URL() (string, bool) {
	if m.Err() != nil {
		return "", false
	}

	e := NewList()
	c := m.Store.Connect(m.User())
	whereDomain := gosql.NewWhere("domain=?", m.Request.Host)
	c.AddModifiers(whereDomain)
	c.Read(&e)
	if c.Err() != nil {
		m.Err(c.Err())
		return "", false
	}

	sort.Sort(e)
	var url string
	var resetCORS bool
	for _, record := range e {
		if strings.HasPrefix(m.Request.URL.Path, record.Path) {
			url = record.URL
			resetCORS = record.ResetCORS
			break
		}
	}

	return url, resetCORS
}

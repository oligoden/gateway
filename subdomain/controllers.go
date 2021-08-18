package subdomain

import (
	"log"
	"net/http"

	"github.com/oligoden/chassis/storage/gosql"
)

func (d Device) Read(subdomain string, url *string, err *error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		m.Bind()

		e := NewRecord()
		c := m.Store.Connect(m.User())
		where := gosql.NewWhere("subdomain=?", subdomain)
		c.AddModifiers(where)
		c.Read(e)
		if c.Err() != nil {
			m.Err(c.Err)
			return
		}

		log.Println(e.URL)
		*url = e.URL
	})
}

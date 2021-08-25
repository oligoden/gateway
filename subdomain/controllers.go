package subdomain

import (
	"log"
	"net/http"
)

func (d Device) Check() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)

		url := m.URL()
		if m.Err() != nil {
			NewView(w).Error(m)
			return
		}

		log.Println("forwarding to", url)
		d.rps[url].ServeHTTP(w, r)
	})
}

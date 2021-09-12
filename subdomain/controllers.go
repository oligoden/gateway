package subdomain

import (
	"log"
	"net/http"
)

func (d Device) Check(dmn string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)

		url := m.URL(dmn)
		if m.Err() != nil {
			NewView(w).Error(m)
			return
		}

		if url == "" {
			log.Println("not forwarding")
			return
		}

		proxy, ok := d.rps[url]
		if !ok {
			log.Println("not forwarding", url, "not registered")
			return
		}

		log.Println("forwarding to", url)
		proxy.ServeHTTP(w, r)
	})
}

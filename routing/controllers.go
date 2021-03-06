package routing

import (
	"log"
	"net/http"
)

func (d Device) Check() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)

		url, resetCORS := m.URL()
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
		if resetCORS {
			w.Header().Del("Access-Control-Allow-Origin")
			w.Header().Del("Access-Control-Allow-Credentials")
		}
		proxy.ServeHTTP(w, r)
	})
}

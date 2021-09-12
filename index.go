package gateway

import (
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

// var (
// 	routerDurations = prometheus.NewHistogram(prometheus.HistogramOpts{
// 		Namespace: "webapp",
// 		Subsystem: "gateway",
// 		Name:      "rerouted_duration_seconds",
// 		Help:      "The duration of rerouted traffic.",
// 		Buckets:   prometheus.LinearBuckets(0.0001, 0.04, 5),
// 	})
// )

type Index struct {
	reverseProxies map[string]*httputil.ReverseProxy
}

func NewIndex() *Index {
	// prometheus.MustRegister(routerDurations)

	h := &Index{}
	h.reverseProxies = make(map[string]*httputil.ReverseProxy)
	return h
}

func (h *Index) SetProxy(key string, p *httputil.ReverseProxy) {
	h.reverseProxies[key] = p
}

func (h *Index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("reverse proxy handler got", r.URL.Path)

	// startTime := time.Now()
	// defer func() {
	// 	endTime := time.Now()
	// 	duration := endTime.Sub(startTime)
	// 	routerDurations.Observe(duration.Seconds())
	// }()

	for k, p := range h.reverseProxies {
		if strings.HasPrefix(r.URL.Path, "/"+k) {
			p.ServeHTTP(w, r)
			return
		}
	}

	log.Println("unhandled path", r.URL.Path)
	w.WriteHeader(http.StatusBadRequest)
}

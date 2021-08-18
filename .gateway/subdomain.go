package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/oligoden/gateway/.gateway/subdomain"
)

type SubDomain struct {
	device *subdomain.Device
}

func NewSubDomain(d *subdomain.Device) *SubDomain {
	h := &SubDomain{}
	h.device = d
	return h
}

func (sd *SubDomain) Check(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "oligoden.com" {
			if h != nil {
				h.ServeHTTP(w, r)
			}
			return
		}
		log.Println("request host", r.Host)

		var subdomain, locationURL string
		var err error

		subdomain = strings.TrimSuffix(r.Host, ".oligoden.com")
		log.Println("request subdomain", subdomain)

		sd.device.Read(subdomain, &locationURL, &err).ServeHTTP(w, r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if locationURL == "" {
			log.Println("empty forward URL")
			fmt.Fprint(w, r.Host, " does not exist")
			return
		}

		log.Println("forward URL", locationURL)
		parsedURL, err := url.Parse(locationURL)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cert, err := tls.LoadX509KeyPair("certs/app01.stg01.gce01.oligoden.com.crt", "certs/app01.stg01.gce01.oligoden.com.key")
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := ioutil.ReadFile("certs/app01.stg01.gce01.oligoden.com.ca.crt")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			log.Fatal("ca cert not added")
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(parsedURL)
		reverseProxy.Transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		}
		reverseProxy.ServeHTTP(w, r)
	})
}

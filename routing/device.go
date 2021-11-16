package routing

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/oligoden/chassis/device"
	"github.com/oligoden/chassis/device/model"
	"github.com/oligoden/chassis/device/view"
)

type Device struct {
	device.Default
	rps map[string]http.Handler
}

func NewDevice(s model.Connector, rpds ...string) *Device {
	d := &Device{}
	nm := func(r *http.Request) model.Operator { return NewModel(r, s) }
	nv := func(w http.ResponseWriter) view.Operator { return NewView(w) }
	d.Default = device.NewDevice(nm, nv, s)

	d.rps = make(map[string]http.Handler)
	caCertPool := x509.NewCertPool()

	for _, c := range strings.Split(os.Getenv("CA_CERTS"), ",") {
		if c == "" {
			break
		}

		caCert, err := ioutil.ReadFile("certs/" + c + ".ca.crt")
		if err != nil {
			log.Fatal(err)
		}
		if !caCertPool.AppendCertsFromPEM(caCert) {
			log.Fatal("ca cert not added")
		}
	}

	for _, c := range strings.Split(os.Getenv("CERTS"), ",") {
		if c == "" {
			break
		}

		cert, err := tls.LoadX509KeyPair("certs/"+c+".crt", "certs/"+c+".key")
		if err != nil {
			log.Fatal(err)
		}

		parsedURL, err := url.Parse("https://" + c)
		if err != nil {
			log.Fatal(err)
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

		d.rps[c] = reverseProxy
	}

	for _, rpd := range strings.Split(os.Getenv("RPDS"), ",") {
		d.setReverseProxy(rpd)
	}

	for _, rpd := range rpds {
		d.setReverseProxy(rpd)
	}

	return d
}

func (d *Device) setReverseProxy(p string) {
	if p == "" {
		return
	}

	parsedURL, err := url.Parse("http://" + p)
	if err != nil {
		log.Fatal(err)
	}

	d.rps[p] = httputil.NewSingleHostReverseProxy(parsedURL)
}

func (d *Device) SetProxy(p string) {
	parsedURL, err := url.Parse("http://" + p)
	if err != nil {
		log.Fatal(err)
	}

	d.rps[p] = httputil.NewSingleHostReverseProxy(parsedURL)
}

func (d *Device) SetProxyHandler(p string, h http.Handler) {
	d.rps[p] = h
}

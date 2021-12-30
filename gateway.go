package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func Serve(mux *http.ServeMux, addr string) {
	if addr == "" {
		addr = ":8080"
	}

	httpServer := &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverError := make(chan error)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverError <- err
			return
		}
		fmt.Println("http server shutdown")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	log.Println("running http server")
	select {
	case err := <-serverError:
		fmt.Println("http server error", err)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case sig := <-quit:
		fmt.Println("\ngot signal", sig)
	}

	shutdown(httpServer)
	time.Sleep(100 * time.Millisecond)
}

func split(r rune) bool {
	return r == ',' || r == ' '
}

func ServeTLS(mux *http.ServeMux) {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(strings.FieldsFunc(os.Getenv("ALLOW"), split)...),
		Email:      os.Getenv("EMAIL"),
	}

	httpServer := &http.Server{
		Addr:           ":80",
		Handler:        certManager.HTTPHandler(nil),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate:           certManager.GetCertificate,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				// Best disabled, as they don't provide Forward Secrecy,
				// but might be necessary for some clients
				// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serverHTTPError := make(chan error)
	serverHTTPSError := make(chan error)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverHTTPError <- err
			return
		}
		fmt.Println("http server shutdown")
	}()
	go func() {
		err := httpsServer.ListenAndServeTLS("", "")
		if err != nil && err != http.ErrServerClosed {
			serverHTTPSError <- err
			return
		}
		fmt.Println("https server shutdown")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	log.Println("running https server")
	select {
	case err := <-serverHTTPError:
		fmt.Println("server error", err)
		shutdown(httpsServer)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case err := <-serverHTTPSError:
		fmt.Println("server error", err)
		shutdown(httpServer)
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	case sig := <-quit:
		fmt.Println("\ngot signal", sig)
	}

	shutdown(httpServer)
	shutdown(httpsServer)
	time.Sleep(100 * time.Millisecond)
}

func shutdown(s *http.Server) {
	ctxServer, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctxServer)
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("https server shutdown error", err)
	}
}

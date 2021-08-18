package main

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

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"
	"golang.org/x/crypto/acme/autocert"

	//---
	"github.com/oligoden/gateway"
	"github.com/oligoden/gateway/session"
	"github.com/oligoden/gateway/subdomain"
	//end
	//+++
	// "github.com/oligoden/gateway/.gateway"
	// "github.com/oligoden/gateway/.gateway/session"
	// "github.com/oligoden/gateway/.gateway/subdomain"
	//end
)

func serveFile(f string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, f)
	})
}

func serveFiles(p, d string) http.Handler {
	return http.StripPrefix(p, http.FileServer(http.Dir(d)))
}

func main() {
	hIndex := gateway.NewIndex()
	// task, _ := url.Parse("http://task:8080/")
	// hIndex.SetProxy("tasks", httputil.NewSingleHostReverseProxy(task))

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbAddr := os.Getenv("DB_ADDRESS")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	params := "charset=utf8&parseTime=True&loc=Local"
	format := "%s:%s@tcp(%s:%s)/%s?%s"

	if dbPort == "" {
		dbPort = "3306"
	}
	uri := fmt.Sprintf(format, dbUser, dbPass, dbAddr, dbPort, dbName, params)
	dbt := "mysql"

	store := gosql.New(dbt, uri)
	if store.Err() != nil {
		log.Fatal(store.Err())
	}

	dSession := session.NewDevice(store)
	store.Migrate(session.NewRecord())
	store.Migrate(session.NewSessionUsersRecord())

	dSubDomain := subdomain.NewDevice(store)
	store.Migrate(subdomain.NewRecord())
	sd := gateway.NewSubDomain(dSubDomain)

	mux := http.NewServeMux()
	mux.Handle("/", adapter.Core(sd.Check(serveFile("static/html"))).Notify().Entry())
	mux.Handle("/static/", adapter.Core(sd.Check(serveFiles("/static/", "static"))).Entry())

	mwProfile := adapter.Core(hIndex).And(dSession.CreateUser())
	mux.Handle("/api/v1/profiles", mwProfile.And(sd.Check(nil)).And(dSession.Authenticate()).Notify().Entry())

	mwSession := adapter.MNA().Delete(dSession.Signout()).Post(dSession.Signin())
	mux.Handle("/api/v1/sessions", mwSession.And(sd.Check(nil)).And(dSession.Authenticate()).Notify().Entry())

	mux.Handle("/api/v1/", adapter.Core(sd.Check(hIndex)).And(dSession.Authenticate()).Notify().Entry())

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(strings.Split(os.Getenv("ALLOW"), ",")...),
		Email:      "info@oligoden.com",
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

	log.Println("running gateway")
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

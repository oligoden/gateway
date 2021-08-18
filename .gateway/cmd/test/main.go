package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"

	 "github.com/oligoden/gateway/.gateway"
	 "github.com/oligoden/gateway/.gateway/session"
	 "github.com/oligoden/gateway/.gateway/subdomain"
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
	// basic, _ := url.Parse("http://basic:8080/")
	// hIndex.SetProxy("basics", httputil.NewSingleHostReverseProxy(basic))

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

	httpServer := &http.Server{
		Addr:           ":8080",
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

	log.Println("running test gateway")
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

func shutdown(s *http.Server) {
	ctxServer, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctxServer)
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("https server shutdown error", err)
	}
}

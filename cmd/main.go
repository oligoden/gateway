package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"

	//---
	"github.com/oligoden/gateway"
	"github.com/oligoden/gateway/routing"
	"github.com/oligoden/gateway/session"
	//end
	//+++
	// "github.com/oligoden/gateway/.gateway"
	// "github.com/oligoden/gateway/.gateway/session"
	// "github.com/oligoden/gateway/.gateway/routing"
	//end
)

func serveFile(f string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("serving file", f)
		http.ServeFile(w, r, f)
	})
}

func serveFiles(p, d string) http.Handler {
	return http.StripPrefix(p, http.FileServer(http.Dir(d)))
}

var Mux func(*adapter.Mux) = func(mux *adapter.Mux) {
	s := mux.Stores["mysqldb"]

	dSession := session.NewDevice(s, mux.URL.Hostname())
	s.Migrate(session.NewRecord())
	s.Migrate(session.NewSessionUsersRecord())

	dRouting := routing.NewDevice(s, mux.RPDs...)
	s.Migrate(routing.NewRecord())

	mux.Handle("/").
		Core(serveFile("static/index.html")).
		SubDomain(dRouting.Check(), "-api").
		And(dSession.Authenticate()).
		Notify().Entry()

	mux.Handle("/static/").
		Core(serveFiles("/static/", "static")).
		And(dSession.Authenticate()).
		Notify().Entry()

	mux.Handle("/profiles").
		NotFound().
		SubDomain(dRouting.Check(), "api").
		And(dSession.CreateUser()).
		And(dSession.Authenticate()).
		CORS().Notify().Entry()

	mux.Handle("/sessions").
		MethodNotAllowed().
		Delete(dSession.Signout()).
		Post(dSession.Signin()).
		SubDomain(dRouting.Check(), "-api").
		And(dSession.Authenticate()).
		CORS().Notify().Entry()
}

func main() {
	store := gosql.New(gosql.ConnURL())
	if store.Err() != nil {
		log.Fatalln("could not connect to store ->", store.Err())
	}

	mux := adapter.NewMux().
		SetURL("http://test.com:8080").
		SetStore("mysqldb", store).
		AddRPD("profile:8080").Compile(Mux)

	gateway.Serve(mux)
}

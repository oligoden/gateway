package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/oligoden/chassis/adapter"
	"github.com/oligoden/chassis/storage/gosql"

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
	domain := "example.com"

	hIndex := gateway.NewIndex()
	// hIndex := gateway.NewIndexTLS()
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
	dSubDomain.SetProxy("test.example.com")
	store.Migrate(subdomain.NewRecord())

	mux := http.NewServeMux()

	mwRoot := adapter.New(domain)
	mwRoot = mwRoot.Core(serveFile("static/html"))
	mwRoot = mwRoot.SubDomain(dSubDomain.Check())
	mwRoot = mwRoot.And(dSession.Authenticate()).Notify()
	mux.Handle("/", mwRoot.Entry())

	mwStatic := adapter.New(domain)
	mwStatic = mwStatic.Core(serveFiles("/static/", "static"))
	mwStatic = mwStatic.SubDomain(dSubDomain.Check(), "!api")
	mwStatic = mwStatic.And(dSession.Authenticate())
	mux.Handle("/static/", mwStatic.Entry())

	mwProfile := adapter.New(domain)
	mwProfile = mwProfile.Core(hIndex)
	mwProfile = mwProfile.SubDomain(dSubDomain.Check(), "api")
	mwProfile = mwProfile.And(dSession.CreateUser())
	mwProfile = mwProfile.SubDomain(dSubDomain.Check(), "!api")
	mwProfile = mwProfile.And(dSession.Authenticate()).Notify()
	mux.Handle("/profiles", mwProfile.Entry())

	mwSession := adapter.New(domain)
	mwSession = mwSession.MNA()
	mwSession = mwSession.Delete(dSession.Signout()).Post(dSession.Signin())
	mwSession = mwSession.SubDomain(dSubDomain.Check(), "!api")
	mwSession = mwSession.And(dSession.Authenticate()).Notify()
	mux.Handle("/sessions", mwSession.Entry())

	gateway.Serve(mux)
	// gateway.ServeTLS(mux)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

var (
	dbt = "mysql"
	uri = ""
)

func main() {
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
	uri = fmt.Sprintf(format, dbUser, dbPass, dbAddr, dbPort, dbName, params)
	dbt = "mysql"

	mux := mux("oligoden.com")
	gateway.Serve(mux)
	// gateway.ServeTLS(mux)
}

func serveFile(f string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("serving file", f)
		http.ServeFile(w, r, f)
	})
}

func serveFiles(p, d string) http.Handler {
	return http.StripPrefix(p, http.FileServer(http.Dir(d)))
}

func mux(dmn string) *http.ServeMux {
	mux := http.NewServeMux()

	hIndex := gateway.NewIndex()
	// hIndex := gateway.NewIndexTLS()
	profile, _ := url.Parse("http://profile/")
	hIndex.SetProxy("profiles", httputil.NewSingleHostReverseProxy(profile))
	// task, _ := url.Parse("http://task:8080/")
	// hIndex.SetProxy("tasks", httputil.NewSingleHostReverseProxy(task))

	store := gosql.New(dbt, uri)
	if store.Err() != nil {
		log.Fatal(store.Err())
	}

	dSession := session.NewDevice(store)
	store.Migrate(session.NewRecord())
	store.Migrate(session.NewSessionUsersRecord())

	dSubDomain := subdomain.NewDevice(store)
	dSubDomain.SetProxyHandler("api."+dmn, hIndex)
	dSubDomain.SetProxy("staging." + dmn)
	store.Migrate(subdomain.NewRecord())

	mux.Handle("/", adapter.New(dmn).
		Core(serveFile("static/index.html")).
		SubDomain(dSubDomain.Check(dmn), "-api").
		And(dSession.Authenticate()).
		Notify().Entry())

	mux.Handle("/static/", adapter.New(dmn).
		Core(serveFiles("/static/", "static")).
		SubDomain(dSubDomain.Check(dmn), "-api").
		And(dSession.Authenticate()).
		Notify().Entry())

	mux.Handle("/profiles", adapter.New(dmn).
		NotFound().
		SubDomain(dSubDomain.Check(dmn), "api").
		And(dSession.CreateUser()).
		SubDomain(dSubDomain.Check(dmn), "-api").
		And(dSession.Authenticate()).
		Notify().Entry())

	mux.Handle("/sessions", adapter.New(dmn).MNA().
		Delete(dSession.Signout()).Post(dSession.Signin()).
		SubDomain(dSubDomain.Check(dmn), "!api").
		And(dSession.Authenticate()).
		Notify().Entry())

	return mux
}

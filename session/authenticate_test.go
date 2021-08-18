package session

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/oligoden/chassis/storage/gosql"
)

const (
	dbt = "mysql"
	uri = "user:pass@tcp(localhost:3308)/test?charset=utf8&parseTime=True&loc=Local"
)

func TestFirstConn(t *testing.T) {
	db := testDBSetup(t)
	defer db.Close()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X_Session_User", "")
	w := httptest.NewRecorder()

	s := gosql.New(dbt, uri)
	if s.Err() != nil {
		t.Error(s.Err().Error())
	}
	s.Migrate(NewRecord())
	d := NewDevice(s)
	d.Authenticate().ServeHTTP(w, r)

	var id uint
	var hash string
	err := db.QueryRow("SELECT id, hash from sessions").Scan(&id, &hash)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(uint(1), id)
	assert.Equal("0", r.Header.Get("X_User"))
	if assert.NotEmpty(w.Result().Cookies()) {
		assert.Equal(hash, w.Result().Cookies()[0].Value)
	}
}

func TestAuthenticateWithCookie(t *testing.T) {
	t.SkipNow()
	qs := []string{
		"INSERT INTO `sessions` (`uc`, `owner_id`, `perms`, `hash`) VALUES ('a', 1, ':::r', 'xyz')",
		"INSERT INTO `sessions` (`uc`, `owner_id`, `perms`, `hash`) VALUES ('b', 1, ':::r', 'tyu')",
		"INSERT INTO `session_users` (`session_id`, `user_id`) VALUES (2, 1)",
		"INSERT INTO `session_users` (`session_id`, `user_id`) VALUES (2, 2)",
		"INSERT INTO `users` (`uc`, `username`, `perms`, `hash`) VALUES ('c', 'usra', ':::r', 'vbn')",
		"INSERT INTO `users` (`uc`, `username`, `perms`, `hash`) VALUES ('d', 'usrb', ':::r', 'ghj')",
	}
	db := testDBSetup(t, qs...)
	defer db.Close()

	// with valid cookie but no user associated the user id = 0
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X_User", "")
	expire := time.Now().Add(24 * 200 * time.Hour)
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "xyz",
		Path:     "/",
		Expires:  expire,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	r.AddCookie(cookie)
	w := httptest.NewRecorder()

	s := gosql.New(dbt, uri)
	if s.Err() != nil {
		t.Error(s.Err())
	}
	d := NewDevice(s)
	d.Authenticate().ServeHTTP(w, r)

	assert := assert.New(t)
	assert.Equal("0", r.Header.Get("X_User"))

	// with valid cookie and users associated but not specified, the first user is selected
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X_User", "")
	cookie = &http.Cookie{
		Name:     "session",
		Value:    "tyu",
		Path:     "/",
		Expires:  expire,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	r.AddCookie(cookie)
	w = httptest.NewRecorder()

	s = gosql.New(dbt, uri)
	d = NewDevice(s)
	d.Authenticate().ServeHTTP(w, r)

	assert.Equal("1", r.Header.Get("X_User"))

	// with valid cookie and users associated and specified, the specific user is selected
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X_User", "usrb")
	r.AddCookie(cookie)
	w = httptest.NewRecorder()

	s = gosql.New(dbt, uri)
	d = NewDevice(s)
	d.Authenticate().ServeHTTP(w, r)

	assert.Equal("2", r.Header.Get("X_User"))

	// with valid cookie and users associated and incorrect specified, the first user is selected
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X_User", "usrc")
	r.AddCookie(cookie)
	w = httptest.NewRecorder()

	s = gosql.New(dbt, uri)
	d = NewDevice(s)
	d.Authenticate().ServeHTTP(w, r)

	assert.Equal("1", r.Header.Get("X_User"))
}

func testDBSetup(t *testing.T, qs ...string) *sql.DB {
	db, err := sql.Open(dbt, uri)
	if err != nil {
		t.Error(err)
	}

	db.Exec("DROP TABLE users")
	db.Exec("DROP TABLE groups")
	db.Exec("DROP TABLE record_groups")
	db.Exec("DROP TABLE record_users")

	db.Exec("DROP TABLE sessions")
	db.Exec("DROP TABLE session_users")

	q := "CREATE TABLE `sessions` (`id` int unsigned AUTO_INCREMENT, `ip` varchar(255) NOT NULL DEFAULT '', `user_agent` varchar(511) NOT NULL DEFAULT '', `uc` varchar(255) UNIQUE, `owner_id` int unsigned, `perms` varchar(255), `hash` varchar(255), PRIMARY KEY (`id`))"
	_, err = db.Exec(q)
	if err != nil {
		t.Fatal(err)
	}

	q = "CREATE TABLE `session_users` (`id` int unsigned AUTO_INCREMENT, `session_id` int unsigned, `user_id` int unsigned, PRIMARY KEY (`id`))"
	_, err = db.Exec(q)
	if err != nil {
		t.Fatal(err)
	}

	for _, q = range qs {
		_, err = db.Exec(q)
		if err != nil {
			t.Fatal(err)
		}
	}

	return db
}

package routing_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oligoden/chassis/storage/gosql"
	"github.com/oligoden/gateway/routing"
	"github.com/stretchr/testify/assert"
)

func TestURLModel(t *testing.T) {
	uri := "user:pass@tcp(localhost:3308)/test?charset=utf8&parseTime=True&loc=Local"

	db := testDBDropTables(t, uri)
	defer db.Close()

	store := gosql.New(uri)
	if store.Err() != nil {
		t.Fatal("could not connect to store")
	}
	store.Migrate(routing.NewRecord())

	qs := []string{
		"INSERT INTO `routings` (`uc`, `domain`, `path`, `url`, `reset_cors`, `owner_id`, `perms`, `hash`) VALUES ('a', 'oligoden.com', '/a/', 'a.com', false, 1, ':::r', 'abc')",
		"INSERT INTO `routings` (`uc`, `domain`, `path`, `url`, `reset_cors`, `owner_id`, `perms`, `hash`) VALUES ('b', 'oligoden.com', '/a/b', 'b.com', false, 1, ':::r', 'fgh')",
		"INSERT INTO `routings` (`uc`, `domain`, `path`, `url`, `reset_cors`, `owner_id`, `perms`, `hash`) VALUES ('c', 'oligoden.com', '/a/c', 'c.com', false, 1, ':::r', 'nbv')",
	}
	testDBSetup(db, t, qs...)

	req := httptest.NewRequest(http.MethodGet, "/a/a", nil)
	req.Host = "oligoden.com"
	req.Header.Set("X_User", "1")
	req.Header.Set("X_Session", "1")

	m := routing.NewModel(req, store)
	path, _ := m.URL()

	assert := assert.New(t)
	assert.NoError(m.Err())
	assert.Equal("a.com", path)

	req = httptest.NewRequest(http.MethodGet, "/a/b", nil)
	req.Host = "oligoden.com"
	req.Header.Set("X_User", "1")
	req.Header.Set("X_Session", "1")

	m = routing.NewModel(req, store)
	path, _ = m.URL()

	assert.NoError(m.Err())
	assert.Equal("b.com", path)

	req = httptest.NewRequest(http.MethodGet, "/c/c", nil)
	req.Host = "oligoden.com"
	req.Header.Set("X_User", "1")
	req.Header.Set("X_Session", "1")

	m = routing.NewModel(req, store)
	path, _ = m.URL()

	assert.NoError(m.Err())
	assert.Equal("", path)
}

func testDBDropTables(t *testing.T, uri string) *sql.DB {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		t.Error(err)
	}

	db.Exec("DROP TABLE users")
	db.Exec("DROP TABLE groups")
	db.Exec("DROP TABLE record_groups")
	db.Exec("DROP TABLE record_users")

	db.Exec("DROP TABLE sessions")
	db.Exec("DROP TABLE session_users")

	db.Exec("DROP TABLE routings")

	return db
}

func testDBSetup(db *sql.DB, t *testing.T, qs ...string) {
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			t.Fatal(err)
		}
	}
}

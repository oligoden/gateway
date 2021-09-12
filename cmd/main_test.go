package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

func TestRoot(t *testing.T) {
	uri = "user:pass@tcp(localhost:3308)/test?charset=utf8&parseTime=True&loc=Local"

	db := testDBDropTables(t)
	defer db.Close()

	mux := mux("oligoden.com")
	qs := []string{
		"INSERT INTO `subdomains` (`uc`, `subdomain`, `url`, `owner_id`, `perms`, `hash`) VALUES ('a', 'staging', 'staging.oligoden.com', 1, ':::r', 'abc')",
	}
	testDBSetup(db, t, qs...)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "oligoden.com"

	apitest.New().
		// Report(apitest.SequenceDiagram()).
		Handler(mux).
		HttpRequest(req).
		Expect(t).
		Status(http.StatusOK).
		End()

	stagingMock := apitest.NewMock().
		Get("http://staging.oligoden.com/").
		RespondWith().
		Body(`<staging html page>`).
		Status(http.StatusOK).
		End()

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "staging.oligoden.com"

	apitest.New().
		// Report(apitest.SequenceDiagram()).
		Mocks(stagingMock).
		Handler(mux).
		HttpRequest(req).
		Expect(t).
		Status(http.StatusOK).
		Body(`<staging html page>`).
		End()
}

func Test(t *testing.T) {
	uri = "user:pass@tcp(localhost:3308)/test?charset=utf8&parseTime=True&loc=Local"

	db := testDBDropTables(t)
	defer db.Close()

	mux := mux("sut")
	qs := []string{
		// "INSERT INTO `sessions` (`uc`, `owner_id`, `perms`, `hash`) VALUES ('a', 1, ':::r', 'xyz')",
		// "INSERT INTO `sessions` (`uc`, `owner_id`, `perms`, `hash`) VALUES ('b', 1, ':::r', 'tyu')",
		// "INSERT INTO `session_users` (`session_id`, `user_id`) VALUES (2, 1)",
		// "INSERT INTO `session_users` (`session_id`, `user_id`) VALUES (2, 2)",
		// "INSERT INTO `users` (`uc`, `username`, `perms`, `hash`) VALUES ('c', 'usra', ':::r', 'vbn')",
		// "INSERT INTO `users` (`uc`, `username`, `perms`, `hash`) VALUES ('d', 'usrb', ':::r', 'ghj')",
	}
	testDBSetup(db, t, qs...)

	profileMock := apitest.NewMock().
		Post("http://example.com/profiles").
		RespondWith().
		Body(`{"name": "jon"}`).
		Status(http.StatusOK).
		End()

	req := httptest.NewRequest(http.MethodPost, "/profiles", nil)
	req.Host = "example.com"

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(profileMock).
		Handler(mux).
		HttpRequest(req).
		Expect(t).
		Status(http.StatusOK).
		End()

	var id uint
	err := db.QueryRow("SELECT owner_id from users").Scan(&id)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(uint(1), id)
}

func testDBDropTables(t *testing.T) *sql.DB {
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

	db.Exec("DROP TABLE subdomains")

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

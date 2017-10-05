package handlers_test

import (
	"database/sql"
	"github.com/aclindsa/moneygo/internal/config"
	"github.com/aclindsa/moneygo/internal/db"
	"github.com/aclindsa/moneygo/internal/handlers"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

var server *httptest.Server

func RunTests(m *testing.M) int {
	tmpdir, err := ioutil.TempDir("./", "handlertest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	dbpath := path.Join(tmpdir, "moneygo.sqlite")
	database, err := sql.Open("sqlite3", "file:"+dbpath+"?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	dbmap, err := db.GetDbMap(database, config.SQLite)
	if err != nil {
		log.Fatal(err)
	}

	servemux := handlers.GetHandler(dbmap)
	server = httptest.NewServer(servemux)
	defer server.Close()

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(RunTests(m))
}

func TestSecurityTemplates(t *testing.T) {
	var sl handlers.SecurityList
	response, err := http.Get(server.URL + "/securitytemplate/?search=USD&type=currency")
	if err != nil {
		t.Error(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}

	err = (&sl).Read(string(body))
	if err != nil {
		t.Error(err)
	}

	num_usd := 0
	for _, s := range *sl.Securities {
		if s.Type != handlers.Currency {
			t.Fatalf("Requested Currency-only security templates, received a non-Currency template for %s", s.Name)
		}

		if s.Name == "USD" && s.AlternateId == "840" {
			num_usd++
		}
	}

	if num_usd != 1 {
		t.Fatalf("Expected one USD security template, found %d\n", num_usd)
	}
}

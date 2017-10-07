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
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
)

var server *httptest.Server

func Delete(client *http.Client, url string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(request)
}

func PutForm(client *http.Client, url string, data url.Values) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return client.Do(request)
}

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
	server = httptest.NewTLSServer(servemux)
	defer server.Close()

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(RunTests(m))
}

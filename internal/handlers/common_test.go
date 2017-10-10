package handlers_test

import (
	"database/sql"
	"encoding/json"
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

type TransactType interface {
	Read(string) error
}

func create(client *http.Client, input, output TransactType, urlsuffix, key string) error {
	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	response, err := client.PostForm(server.URL+urlsuffix, url.Values{key: {string(bytes)}})
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return &e
	}

	err = output.Read(string(body))
	if err != nil {
		return err
	}

	return nil
}

func read(client *http.Client, output TransactType, urlsuffix, key string) error {
	response, err := client.Get(server.URL + urlsuffix)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return &e
	}

	err = output.Read(string(body))
	if err != nil {
		return err
	}

	return nil
}

func update(client *http.Client, input, output TransactType, urlsuffix, key string) error {
	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	response, err := PutForm(client, server.URL+urlsuffix, url.Values{key: {string(bytes)}})
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return &e
	}

	err = output.Read(string(body))
	if err != nil {
		return err
	}

	return nil
}

func remove(client *http.Client, urlsuffix, key string) error {
	response, err := Delete(client, server.URL+urlsuffix)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return &e
	}

	return nil
}

func RunWith(t *testing.T, d *TestData, fn TestDataFunc) {
	testdata, err := d.Initialize()
	if err != nil {
		t.Fatal("Failed to initialize test data: %s", err)
	}
	defer testdata.Teardown()

	fn(t, testdata)
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

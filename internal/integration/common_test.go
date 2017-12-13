package integration_test

import (
	"bytes"
	"encoding/json"
	"github.com/aclindsa/moneygo/internal/config"
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store/db"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
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

func Put(client *http.Client, url string, contentType string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	return client.Do(request)
}

type TransactType interface {
	Read(string) error
}

func create(client *http.Client, input, output TransactType, urlsuffix string) error {
	obj, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}
	response, err := client.Post(server.URL+urlsuffix, "application/json", bytes.NewReader(obj))
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

func read(client *http.Client, output TransactType, urlsuffix string) error {
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

func update(client *http.Client, input, output TransactType, urlsuffix string) error {
	obj, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err
	}
	response, err := Put(client, server.URL+urlsuffix, "application/json", bytes.NewReader(obj))
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

func remove(client *http.Client, urlsuffix string) error {
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

func uploadFile(client *http.Client, filename, urlsuffix string) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	filewriter, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(filewriter, file); err != nil {
		return err
	}

	mw.Close()

	response, err := client.Post(server.URL+urlsuffix, mw.FormDataContentType(), &buf)
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

func NewAmount(amt string) models.Amount {
	var a models.Amount
	if _, ok := a.SetString(amt); !ok {
		panic("Unable to call Amount.SetString()")
	}
	return a
}

func amountsMatch(a models.Amount, amt string) bool {
	cmp := NewAmount(amt)
	return a.Cmp(&cmp.Rat) == 0
}

func accountBalanceHelper(t *testing.T, client *http.Client, account *models.Account, balance string) {
	t.Helper()
	transactions, err := getAccountTransactions(client, account.AccountId, 0, 0, "")
	if err != nil {
		t.Fatalf("Couldn't fetch account transactions for '%s': %s\n", account.Name, err)
	}

	if !amountsMatch(transactions.EndingBalance, balance) {
		t.Errorf("Expected ending balance for '%s' to be '%s', but found %s\n", account.Name, balance, transactions.EndingBalance)
	}
}

func RunWith(t *testing.T, d *TestData, fn TestDataFunc) {
	testdata, err := d.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize test data: %s", err)
	}
	defer func() {
		err := testdata.Teardown()
		if err != nil {
			t.Fatal(err)
		}
	}()

	fn(t, testdata)
}

func RunTests(m *testing.M) int {
	envDbType := os.Getenv("MONEYGO_TEST_DB")
	var dbType config.DbType
	var dsn string

	switch envDbType {
	case "", "sqlite", "sqlite3":
		dbType = config.SQLite
		dsn = ":memory:"
	case "mariadb", "mysql":
		dbType = config.MySQL
		dsn = "root@127.0.0.1/moneygo_test&parseTime=true"
	case "postgres", "postgresql":
		dbType = config.Postgres
		dsn = "postgres://postgres@localhost/moneygo_test"
	default:
		log.Fatalf("Invalid value for $MONEYGO_TEST_DB: %s\n", envDbType)
	}

	if envDSN := os.Getenv("MONEYGO_TEST_DSN"); len(envDSN) > 0 {
		dsn = envDSN
	}

	db, err := db.GetStore(dbType, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Empty() // clear the DB tables

	server = httptest.NewTLSServer(&handlers.APIHandler{Store: db})
	defer server.Close()

	return m.Run()
}

func TestMain(m *testing.M) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	os.Exit(RunTests(m))
}

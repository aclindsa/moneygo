package integration_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"io/ioutil"
	"testing"
)

func TestSecurityTemplates(t *testing.T) {
	var sl models.SecurityList
	response, err := server.Client().Get(server.URL + "/v1/securitytemplates/?search=USD&type=currency")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("Unexpected HTTP status code: %d\n", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = (&sl).Read(string(body))
	if err != nil {
		t.Fatal(err)
	}

	num_usd := 0
	if sl.Securities != nil {
		for _, s := range *sl.Securities {
			if s.Type != models.Currency {
				t.Fatalf("Requested Currency-only security templates, received a non-Currency template for %s", s.Name)
			}

			if s.Name == "USD" && s.AlternateId == "840" {
				num_usd++
			}
		}
	}

	if num_usd != 1 {
		t.Fatalf("Expected one USD security template, found %d\n", num_usd)
	}
}

func TestSecurityTemplateLimit(t *testing.T) {
	var sl models.SecurityList
	response, err := server.Client().Get(server.URL + "/v1/securitytemplates/?search=e&limit=5")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("Unexpected HTTP status code: %d\n", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = (&sl).Read(string(body))
	if err != nil {
		t.Fatal(err)
	}

	if sl.Securities == nil {
		t.Fatalf("Securities was unexpectedly nil\n")
	}

	if len(*sl.Securities) > 5 {
		t.Fatalf("Requested only 5 securities, received %d\n", len(*sl.Securities))
	}
}

func TestSecurityTemplateInvalidType(t *testing.T) {
	var e handlers.Error
	response, err := server.Client().Get(server.URL + "/v1/securitytemplates/?search=e&type=blah")
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = (&e).Read(string(body))
	if err != nil {
		t.Fatal(err)
	}

	if e.ErrorId != 3 {
		t.Fatal("Expected ErrorId 3, Invalid Request")
	}
}

func TestSecurityTemplateInvalidLimit(t *testing.T) {
	var e handlers.Error
	response, err := server.Client().Get(server.URL + "/v1/securitytemplates/?search=e&type=Currency&limit=foo")
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = (&e).Read(string(body))
	if err != nil {
		t.Fatal(err)
	}

	if e.ErrorId != 3 {
		t.Fatal("Expected ErrorId 3, Invalid Request")
	}
}

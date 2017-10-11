package handlers_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
)

func createSecurity(client *http.Client, security *handlers.Security) (*handlers.Security, error) {
	var s handlers.Security
	err := create(client, security, &s, "/security/", "security")
	return &s, err
}

func getSecurity(client *http.Client, securityid int64) (*handlers.Security, error) {
	var s handlers.Security
	err := read(client, &s, "/security/"+strconv.FormatInt(securityid, 10), "security")
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func updateSecurity(client *http.Client, security *handlers.Security) (*handlers.Security, error) {
	var s handlers.Security
	err := update(client, security, &s, "/security/"+strconv.FormatInt(security.SecurityId, 10), "security")
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func deleteSecurity(client *http.Client, s *handlers.Security) error {
	err := remove(client, "/security/"+strconv.FormatInt(s.SecurityId, 10), "security")
	if err != nil {
		return err
	}
	return nil
}

func TestCreateSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			s := d.securities[i]

			if s.SecurityId == 0 {
				t.Errorf("Unable to create security: %+v", s)
			}
			if s.Name != orig.Name {
				t.Errorf("Name doesn't match")
			}
			if s.Description != orig.Description {
				t.Errorf("Description doesn't match")
			}
			if s.Symbol != orig.Symbol {
				t.Errorf("Symbol doesn't match")
			}
			if s.Precision != orig.Precision {
				t.Errorf("Precision doesn't match")
			}
			if s.Type != orig.Type {
				t.Errorf("Type doesn't match")
			}
			if s.AlternateId != orig.AlternateId {
				t.Errorf("AlternateId doesn't match")
			}
		}
	})
}

func TestGetSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			s, err := getSecurity(d.clients[orig.UserId], curr.SecurityId)
			if err != nil {
				t.Fatalf("Error fetching security: %s\n", err)
			}
			if s.SecurityId != curr.SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if s.Name != orig.Name {
				t.Errorf("Name doesn't match")
			}
			if s.Description != orig.Description {
				t.Errorf("Description doesn't match")
			}
			if s.Symbol != orig.Symbol {
				t.Errorf("Symbol doesn't match")
			}
			if s.Precision != orig.Precision {
				t.Errorf("Precision doesn't match")
			}
			if s.Type != orig.Type {
				t.Errorf("Type doesn't match")
			}
			if s.AlternateId != orig.AlternateId {
				t.Errorf("AlternateId doesn't match")
			}
		}
	})
}

func TestUpdateSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			curr.Name = "EUR"
			curr.Description = "Euro"
			curr.Symbol = "€"
			curr.AlternateId = "978"

			s, err := updateSecurity(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error updating security: %s\n", err)
			}

			if s.SecurityId != curr.SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if s.Name != curr.Name {
				t.Errorf("Name doesn't match")
			}
			if s.Description != curr.Description {
				t.Errorf("Description doesn't match")
			}
			if s.Symbol != curr.Symbol {
				t.Errorf("Symbol doesn't match")
			}
			if s.Precision != curr.Precision {
				t.Errorf("Precision doesn't match")
			}
			if s.Type != curr.Type {
				t.Errorf("Type doesn't match")
			}
			if s.AlternateId != curr.AlternateId {
				t.Errorf("AlternateId doesn't match")
			}
		}
	})
}

func TestDeleteSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			err := deleteSecurity(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error deleting security: %s\n", err)
			}

			_, err = getSecurity(d.clients[orig.UserId], curr.SecurityId)
			if err == nil {
				t.Fatalf("Expected error fetching deleted security")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error fetching deleted security: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error fetching deleted security")
			}
		}
	})
}

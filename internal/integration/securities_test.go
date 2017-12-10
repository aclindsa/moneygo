package integration_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"strconv"
	"testing"
)

func createSecurity(client *http.Client, security *models.Security) (*models.Security, error) {
	var s models.Security
	err := create(client, security, &s, "/v1/securities/")
	return &s, err
}

func getSecurity(client *http.Client, securityid int64) (*models.Security, error) {
	var s models.Security
	err := read(client, &s, "/v1/securities/"+strconv.FormatInt(securityid, 10))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func getSecurities(client *http.Client) (*models.SecurityList, error) {
	var sl models.SecurityList
	err := read(client, &sl, "/v1/securities/")
	if err != nil {
		return nil, err
	}
	return &sl, nil
}

func updateSecurity(client *http.Client, security *models.Security) (*models.Security, error) {
	var s models.Security
	err := update(client, security, &s, "/v1/securities/"+strconv.FormatInt(security.SecurityId, 10))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func deleteSecurity(client *http.Client, s *models.Security) error {
	err := remove(client, "/v1/securities/"+strconv.FormatInt(s.SecurityId, 10))
	if err != nil {
		return err
	}
	return nil
}

func TestCreateSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].securities); i++ {
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
		for i := 0; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			s, err := getSecurity(d.clients[orig.UserId], curr.SecurityId)
			if err != nil {
				t.Fatalf("Error fetching security: %s\n", err)
			}
			if s.SecurityId != curr.SecurityId {
				t.Errorf("SecurityId doesn't match %+v %+v", s, curr)
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

func TestGetSecurities(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		sl, err := getSecurities(d.clients[0])
		if err != nil {
			t.Fatalf("Error fetching securities: %s\n", err)
		}

		numsecurities := 0
		foundIds := make(map[int64]bool)
		for i := 0; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			if curr.UserId != d.users[0].UserId {
				continue
			}
			numsecurities += 1

			found := false
			for _, s := range *sl.Securities {
				if orig.Name == s.Name && orig.Description == s.Description && orig.Symbol == orig.Symbol && orig.Precision == s.Precision && orig.Type == s.Type && orig.AlternateId == s.AlternateId {
					if _, ok := foundIds[s.SecurityId]; ok {
						continue
					}
					foundIds[s.SecurityId] = true
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unable to find matching security: %+v", curr)
			}
		}

		if numsecurities+1 == len(*sl.Securities) {
			for _, s := range *sl.Securities {
				if _, ok := foundIds[s.SecurityId]; !ok {
					if s.SecurityId == d.users[0].DefaultCurrency {
						t.Fatalf("Extra security wasn't default currency, seems like an extra security was created")
					}
					break
				}
			}
		} else if numsecurities != len(*sl.Securities) {
			t.Fatalf("Expected %d securities, received %d", numsecurities, len(*sl.Securities))
		}
	})
}

func TestUpdateSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].securities); i++ {
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
	Outer:
		for i := 0; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			for _, a := range d.accounts {
				if a.SecurityId == curr.SecurityId {
					continue Outer
				}
			}

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

func TestDontDeleteSecurity(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
	Outer:
		for i := 0; i < len(data[0].securities); i++ {
			orig := data[0].securities[i]
			curr := d.securities[i]

			for _, a := range d.accounts {
				if a.SecurityId != curr.SecurityId {
					continue Outer
				}
			}

			err := deleteSecurity(d.clients[orig.UserId], &curr)
			if err == nil {
				t.Fatalf("Expected error deleting in-use security")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 7 { // In Use Error
					t.Fatalf("Unexpected API error deleting in-use security: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error deleting in-use security")
			}
		}
	})
}

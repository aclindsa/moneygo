package integration_test

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func newSession(user *User) (*http.Client, error) {
	var u User

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: nil})
	if err != nil {
		return nil, err
	}

	var client http.Client
	client = *server.Client()
	client.Jar = jar

	create(&client, user, &u, "/v1/sessions/")

	return &client, nil
}

func getSession(client *http.Client) (*models.Session, error) {
	var s models.Session
	err := read(client, &s, "/v1/sessions/")
	return &s, err
}

func deleteSession(client *http.Client) error {
	return remove(client, "/v1/sessions/")
}

func sessionExistsOrError(c *http.Client) error {

	url, err := url.Parse(server.URL)
	if err != nil {
		return err
	}
	cookies := c.Jar.Cookies(url)

	var found_session bool = false
	for _, cookie := range cookies {
		if cookie.Name == "moneygo-session" {
			found_session = true
		}
	}
	if found_session {
		return nil
	}
	return fmt.Errorf("Didn't find 'moneygo-session' cookie in CookieJar")
}

func TestCreateSession(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		if err := sessionExistsOrError(d.clients[0]); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGetSession(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		session, err := getSession(d.clients[0])
		if err != nil {
			t.Fatal(err)
		}

		if len(session.SessionSecret) != 0 {
			t.Error("Session.SessionSecret should not be passed back in JSON")
		}

		if session.UserId != d.users[0].UserId {
			t.Errorf("session's UserId (%d) should equal user's UserID (%d)", session.UserId, d.users[0].UserId)
		}

		if session.SessionId == 0 {
			t.Error("session's SessionId should not be 0")
		}
	})
}

func TestDeleteSession(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		err := deleteSession(d.clients[0])
		if err != nil {
			t.Fatalf("Unexpected error removing session: %s\n", err)
		}
		err = deleteSession(d.clients[0])
		if err != nil {
			t.Fatalf("Unexpected error attempting to delete nonexistent session: %s\n", err)
		}
		_, err = getSession(d.clients[0])
		if err == nil {
			t.Fatalf("Expected error fetching deleted session")
		}
		if herr, ok := err.(*handlers.Error); ok {
			if herr.ErrorId != 1 { // Not Signed in
				t.Fatalf("Unexpected API error fetching deleted session: %s", herr)
			}
		} else {
			t.Fatalf("Unexpected error fetching deleted session")
		}

		// Login again so we don't screw up the TestData teardown code
		userWithPassword := d.users[0]
		userWithPassword.Password = data[0].users[0].Password

		client, err := newSession(&userWithPassword)
		if err != nil {
			t.Fatalf("Unexpected error re-creating session: %s\n", err)
		}
		d.clients[0] = client
	})
}

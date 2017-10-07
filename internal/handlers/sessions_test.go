package handlers_test

import (
	"encoding/json"
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func newSession(user *User) (*http.Client, error) {
	var u User
	var e handlers.Error

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: nil})
	if err != nil {
		return nil, err
	}

	client := server.Client()
	client.Jar = jar

	bytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	response, err := client.PostForm(server.URL+"/session/", url.Values{"user": {string(bytes)}})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	err = (&u).Read(string(body))
	if err != nil {
		return nil, err
	}

	err = (&e).Read(string(body))
	if err != nil {
		return nil, err
	}

	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return nil, fmt.Errorf("Unexpected error when creating session %+v", e)
	}

	return client, nil
}

func getSession(client *http.Client) (*handlers.Session, error) {
	var s handlers.Session
	response, err := client.Get(server.URL + "/session/")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	err = (&s).Read(string(body))
	if err != nil {
		return nil, err
	}

	return &s, nil
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
	u, err := createUser(&users[0])
	if err != nil {
		t.Fatal(err)
	}
	u.Password = users[0].Password

	client, err := newSession(u)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteUser(client, u)
	if err := sessionExistsOrError(client); err != nil {
		t.Fatal(err)
	}
}

func TestGetSession(t *testing.T) {
	u, err := createUser(&users[0])
	if err != nil {
		t.Fatal(err)
	}
	u.Password = users[0].Password

	client, err := newSession(u)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteUser(client, u)
	session, err := getSession(client)
	if err != nil {
		t.Fatal(err)
	}

	if len(session.SessionSecret) != 0 {
		t.Error("Session.SessionSecret should not be passed back in JSON")
	}

	if session.UserId != u.UserId {
		t.Errorf("session's UserId (%d) should equal user's UserID (%d)", session.UserId, u.UserId)
	}

	if session.SessionId == 0 {
		t.Error("session's SessionId should not be 0")
	}
}

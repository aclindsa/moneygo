package handlers_test

import (
	"encoding/json"
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func createUser(user *User) (*User, error) {
	bytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	response, err := server.Client().PostForm(server.URL+"/user/", url.Values{"user": {string(bytes)}})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return nil, err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return nil, fmt.Errorf("Error when creating user %+v", e)
	}

	var u User
	err = (&u).Read(string(body))
	if err != nil {
		return nil, err
	}

	if u.UserId == 0 || len(u.Username) == 0 {
		return nil, fmt.Errorf("Unable to create user: %+v", user)
	}

	return &u, nil
}

func updateUser(client *http.Client, user *User) (*User, error) {
	bytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	response, err := PutForm(client, server.URL+"/user/"+strconv.FormatInt(user.UserId, 10), url.Values{"user": {string(bytes)}})
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return nil, err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return nil, fmt.Errorf("Error when updating user %+v", e)
	}

	var u User
	err = (&u).Read(string(body))
	if err != nil {
		return nil, err
	}

	if u.UserId == 0 || len(u.Username) == 0 {
		return nil, fmt.Errorf("Unable to update user: %+v", user)
	}

	return &u, nil
}

func deleteUser(client *http.Client, u *User) error {
	response, err := Delete(client, server.URL+"/user/"+strconv.FormatInt(u.UserId, 10))
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
		return fmt.Errorf("Error when deleting user %+v", e)
	}

	return nil
}

func getUser(client *http.Client, userid int64) (*User, error) {
	response, err := client.Get(server.URL + "/user/" + strconv.FormatInt(userid, 10))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return nil, err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return nil, fmt.Errorf("Error when get user %+v", e)
	}

	var u User
	err = (&u).Read(string(body))
	if err != nil {
		return nil, err
	}

	if u.UserId == 0 || len(u.Username) == 0 {
		return nil, fmt.Errorf("Unable to get userid: %d", userid)
	}

	return &u, nil
}

func TestCreateUser(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		if len(d.users[0].Password) != 0 || len(d.users[0].PasswordHash) != 0 {
			t.Error("Never send password, only send password hash when necessary")
		}
	})
}

func TestGetUser(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		u, err := getUser(d.clients[0], d.users[0].UserId)
		if err != nil {
			t.Fatalf("Error fetching user: %s\n", err)
		}
		if u.UserId != d.users[0].UserId {
			t.Errorf("UserId doesn't match")
		}
	})
}

func TestUpdateUser(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		user := &d.users[0]
		user.Name = "Bob"
		user.Email = "bob@example.com"

		u, err := updateUser(d.clients[0], user)
		if err != nil {
			t.Fatalf("Error updating user: %s\n", err)
		}
		if u.UserId != user.UserId {
			t.Errorf("UserId doesn't match")
		}
		if u.Name != user.Name {
			t.Errorf("UserId doesn't match")
		}
		if u.Email != user.Email {
			t.Errorf("UserId doesn't match")
		}
	})
}

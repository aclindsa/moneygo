package integration_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
)

func createUser(user *User) (*User, error) {
	var u User
	err := create(server.Client(), user, &u, "/v1/users/")
	return &u, err
}

func getUser(client *http.Client, userid int64) (*User, error) {
	var u User
	err := read(client, &u, "/v1/users/"+strconv.FormatInt(userid, 10))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func updateUser(client *http.Client, user *User) (*User, error) {
	var u User
	err := update(client, user, &u, "/v1/users/"+strconv.FormatInt(user.UserId, 10))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func deleteUser(client *http.Client, u *User) error {
	err := remove(client, "/v1/users/"+strconv.FormatInt(u.UserId, 10))
	if err != nil {
		return err
	}
	return nil
}

func TestCreateUser(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		if d.users[0].UserId == 0 || len(d.users[0].Username) == 0 {
			t.Errorf("Unable to create user: %+v", data[0].users[0])
		}

		if len(d.users[0].Password) != 0 || len(d.users[0].PasswordHash) != 0 {
			t.Error("Never send password, only send password hash when necessary")
		}
	})
}

func TestDontRecreateUser(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for _, user := range data[0].users {
			_, err := createUser(&user)
			if err == nil {
				t.Fatalf("Expected error re-creating user")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 4 { // User exists
					t.Fatalf("Unexpected API error re-creating user: %s", herr)
				}
			} else {
				t.Fatalf("Expected error re-creating user")
			}
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
		if len(u.Username) == 0 {
			t.Fatalf("Empty username for: %d", d.users[0].UserId)
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
		if u.Username != u.Username {
			t.Errorf("Username doesn't match")
		}
		if u.Name != user.Name {
			t.Errorf("Name doesn't match")
		}
		if u.Email != user.Email {
			t.Errorf("Email doesn't match")
		}
	})
}

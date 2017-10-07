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
	u, err := createUser(&users[0])
	if err != nil {
		t.Fatal(err)
	}

	if len(u.Password) != 0 || len(u.PasswordHash) != 0 {
		t.Error("Never send password, only send password hash when necessary")
	}

	u.Password = users[0].Password

	client, err := newSession(u)
	if err != nil {
		t.Fatalf("Error creating new session, user not deleted (may cause errors in other tests): %s", err)
	}
	defer deleteUser(client, u)
}

func TestGetUser(t *testing.T) {
	origu, err := createUser(&users[0])
	if err != nil {
		t.Fatal(err)
	}
	origu.Password = users[0].Password

	client, err := newSession(origu)
	if err != nil {
		t.Fatalf("Error creating new session, user not deleted (may cause errors in other tests): %s", err)
	}
	defer deleteUser(client, origu)

	u, err := getUser(client, origu.UserId)
	if err != nil {
		t.Fatalf("Error fetching user: %s\n", err)
	}
	if u.UserId != origu.UserId {
		t.Errorf("UserId doesn't match")
	}
}

func TestUpdateUser(t *testing.T) {
	origu, err := createUser(&users[0])
	if err != nil {
		t.Fatal(err)
	}
	origu.Password = users[0].Password

	client, err := newSession(origu)
	if err != nil {
		t.Fatalf("Error creating new session, user not deleted (may cause errors in other tests): %s", err)
	}
	defer deleteUser(client, origu)

	origu.Name = "Bob"
	origu.Email = "bob@example.com"

	u, err := updateUser(client, origu)
	if err != nil {
		t.Fatalf("Error updating user: %s\n", err)
	}
	if u.UserId != origu.UserId {
		t.Errorf("UserId doesn't match")
	}
	if u.Name != origu.Name {
		t.Errorf("UserId doesn't match")
	}
	if u.Email != origu.Email {
		t.Errorf("UserId doesn't match")
	}
}

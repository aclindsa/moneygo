package handlers_test

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Needed because handlers.User doesn't allow Password to be written to JSON

type User struct {
	UserId          int64
	DefaultCurrency int64 // SecurityId of default currency, or ISO4217 code for it if creating new user
	Name            string
	Username        string
	Password        string
	PasswordHash    string
	Email           string
}

func (u *User) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(u)
}

func (u *User) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(u)
}

var users = []User{
	User{
		DefaultCurrency: 840, // USD
		Name:            "John Smith",
		Username:        "jsmith",
		Password:        "hunter2",
		Email:           "jsmith@example.com",
	},
}

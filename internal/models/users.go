package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type User struct {
	UserId          int64
	DefaultCurrency int64 // SecurityId of default currency, or ISO4217 code for it if creating new user
	Name            string
	Username        string
	Password        string `db:"-"`
	PasswordHash    string `json:"-"`
	Email           string
}

const BogusPassword = "password"

func (u *User) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(u)
}

func (u *User) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(u)
}

func (u *User) HashPassword() {
	password_hasher := sha256.New()
	io.WriteString(password_hasher, u.Password)
	u.PasswordHash = fmt.Sprintf("%x", password_hasher.Sum(nil))
	u.Password = ""
}

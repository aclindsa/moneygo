package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/gorp.v1"
	"io"
	"log"
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

type UserExistsError struct{}

func (ueu UserExistsError) Error() string {
	return "User exists"
}

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

func GetUser(db *DB, userid int64) (*User, error) {
	var u User

	err := db.SelectOne(&u, "SELECT * from users where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserTx(transaction *gorp.Transaction, userid int64) (*User, error) {
	var u User

	err := transaction.SelectOne(&u, "SELECT * from users where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByUsername(db *DB, username string) (*User, error) {
	var u User

	err := db.SelectOne(&u, "SELECT * from users where Username=?", username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func InsertUser(db *DB, u *User) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	security_template := FindCurrencyTemplate(u.DefaultCurrency)
	if security_template == nil {
		transaction.Rollback()
		return errors.New("Invalid ISO4217 Default Currency")
	}

	existing, err := transaction.SelectInt("SELECT count(*) from users where Username=?", u.Username)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if existing > 0 {
		transaction.Rollback()
		return UserExistsError{}
	}

	err = transaction.Insert(u)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// Copy the security template and give it our new UserId
	var security Security
	security = *security_template
	security.UserId = u.UserId

	err = InsertSecurityTx(transaction, &security)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// Update the user's DefaultCurrency to our new SecurityId
	u.DefaultCurrency = security.SecurityId
	count, err := transaction.Update(u)
	if err != nil {
		transaction.Rollback()
		return err
	} else if count != 1 {
		transaction.Rollback()
		return errors.New("Would have updated more than one user")
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func GetUserFromSession(db *DB, r *http.Request) (*User, error) {
	s, err := GetSession(db, r)
	if err != nil {
		return nil, err
	}
	return GetUser(db, s.UserId)
}

func UpdateUser(db *DB, u *User) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	security, err := GetSecurityTx(transaction, u.DefaultCurrency, u.UserId)
	if err != nil {
		transaction.Rollback()
		return err
	} else if security.UserId != u.UserId || security.SecurityId != u.DefaultCurrency {
		transaction.Rollback()
		return errors.New("UserId and DefaultCurrency don't match the fetched security")
	} else if security.Type != Currency {
		transaction.Rollback()
		return errors.New("New DefaultCurrency security is not a currency")
	}

	count, err := transaction.Update(u)
	if err != nil {
		transaction.Rollback()
		return err
	} else if count != 1 {
		transaction.Rollback()
		return errors.New("Would have updated more than one user")
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func UserHandler(w http.ResponseWriter, r *http.Request, db *DB) {
	if r.Method == "POST" {
		user_json := r.PostFormValue("user")
		if user_json == "" {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		var user User
		err := user.Read(user_json)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		user.UserId = -1
		user.HashPassword()

		err = InsertUser(db, &user)
		if err != nil {
			if _, ok := err.(UserExistsError); ok {
				WriteError(w, 4 /*User Exists*/)
			} else {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
			}
			return
		}

		w.WriteHeader(201 /*Created*/)
		err = user.Write(w)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	} else {
		user, err := GetUserFromSession(db, r)
		if err != nil {
			WriteError(w, 1 /*Not Signed In*/)
			return
		}

		userid, err := GetURLID(r.URL.Path)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		if userid != user.UserId {
			WriteError(w, 2 /*Unauthorized Access*/)
			return
		}

		if r.Method == "GET" {
			err = user.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else if r.Method == "PUT" {
			user_json := r.PostFormValue("user")
			if user_json == "" {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			// Save old PWHash in case the new password is bogus
			old_pwhash := user.PasswordHash

			err = user.Read(user_json)
			if err != nil || user.UserId != userid {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			// If the user didn't create a new password, keep their old one
			if user.Password != BogusPassword {
				user.HashPassword()
			} else {
				user.Password = ""
				user.PasswordHash = old_pwhash
			}

			err = UpdateUser(db, user)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			err = user.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else if r.Method == "DELETE" {
			count, err := db.Delete(&user)
			if count != 1 || err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}

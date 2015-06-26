package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
)

var cookie_store = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))

type Session struct {
	SessionId     int64
	SessionSecret string `json:"-"`
	UserId        int64
}

func (s *Session) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func GetSession(r *http.Request) (*Session, error) {
	var s Session

	session, _ := cookie_store.Get(r, "moneygo")
	_, ok := session.Values["session-secret"]
	if !ok {
		return nil, fmt.Errorf("session-secret cookie not set")
	}
	s.SessionSecret = session.Values["session-secret"].(string)

	err := DB.SelectOne(&s, "SELECT * from sessions where SessionSecret=?", s.SessionSecret)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func DeleteSessionIfExists(r *http.Request) {
	session, err := GetSession(r)
	if err == nil {
		DB.Delete(session)
	}
}

func NewSession(w http.ResponseWriter, r *http.Request, userid int64) (*Session, error) {
	s := Session{}

	session, _ := cookie_store.Get(r, "moneygo")

	session.Values["session-secret"] = string(securecookie.GenerateRandomKey(64))
	s.SessionSecret = session.Values["session-secret"].(string)
	s.UserId = userid

	err := DB.Insert(&s)
	if err != nil {
		return nil, err
	}

	err = session.Save(r, w)
	if err != nil {
		return nil, err
	} else {
		return &s, nil
	}
}

func SessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" || r.Method == "PUT" {
		user_json := r.PostFormValue("user")
		if user_json == "" {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		user := User{}
		err := user.Read(user_json)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		dbuser, err := GetUserByUsername(user.Username)
		if err != nil {
			WriteError(w, 2 /*Unauthorized Access*/)
			return
		}

		user.HashPassword()
		if user.PasswordHash != dbuser.PasswordHash {
			WriteError(w, 2 /*Unauthorized Access*/)
			return
		}

		DeleteSessionIfExists(r)

		_, err = NewSession(w, r, dbuser.UserId)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			return
		}

		WriteSuccess(w)
	} else if r.Method == "GET" {
		s, err := GetSession(r)
		if err != nil {
			WriteError(w, 1 /*Not Signed In*/)
			return
		}

		s.Write(w)
	} else if r.Method == "DELETE" {
		DeleteSessionIfExists(r)
		WriteSuccess(w)
	}
}

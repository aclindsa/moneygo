package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Session struct {
	SessionId     int64
	SessionSecret string `json:"-"`
	UserId        int64
	Created       time.Time
	Expires       time.Time
}

func (s *Session) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func (s *Session) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(s)
}

func GetSession(tx *Tx, r *http.Request) (*Session, error) {
	var s Session

	cookie, err := r.Cookie("moneygo-session")
	if err != nil {
		return nil, fmt.Errorf("moneygo-session cookie not set")
	}
	s.SessionSecret = cookie.Value

	err = tx.SelectOne(&s, "SELECT * from sessions where SessionSecret=?", s.SessionSecret)
	if err != nil {
		return nil, err
	}

	if s.Expires.Before(time.Now()) {
		tx.Delete(&s)
		return nil, fmt.Errorf("Session has expired")
	}
	return &s, nil
}

func DeleteSessionIfExists(tx *Tx, r *http.Request) error {
	session, err := GetSession(tx, r)
	if err == nil {
		_, err := tx.Delete(session)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewSessionCookie() (string, error) {
	bits := make([]byte, 128)
	if _, err := io.ReadFull(rand.Reader, bits); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bits), nil
}

type NewSessionWriter struct {
	session *Session
	cookie  *http.Cookie
}

func (n *NewSessionWriter) Write(w http.ResponseWriter) error {
	http.SetCookie(w, n.cookie)
	return n.session.Write(w)
}

func NewSession(tx *Tx, r *http.Request, userid int64) (*NewSessionWriter, error) {
	s := Session{}

	session_secret, err := NewSessionCookie()
	if err != nil {
		return nil, err
	}

	existing, err := tx.SelectInt("SELECT count(*) from sessions where SessionSecret=?", session_secret)
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, fmt.Errorf("%d session(s) exist with the generated session_secret", existing)
	}

	cookie := http.Cookie{
		Name:     "moneygo-session",
		Value:    session_secret,
		Path:     "/",
		Domain:   r.URL.Host,
		Expires:  time.Now().AddDate(0, 1, 0), // a month from now
		Secure:   false,
		HttpOnly: true,
	}

	s.SessionSecret = session_secret
	s.UserId = userid
	s.Created = time.Now()
	s.Expires = cookie.Expires

	err = tx.Insert(&s)
	if err != nil {
		return nil, err
	}
	return &NewSessionWriter{&s, &cookie}, nil
}

func SessionHandler(r *http.Request, context *Context) ResponseWriterWriter {
	if r.Method == "POST" || r.Method == "PUT" {
		var user User
		if err := ReadJSON(r, &user); err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		dbuser, err := GetUserByUsername(context.Tx, user.Username)
		if err != nil {
			return NewError(2 /*Unauthorized Access*/)
		}

		user.HashPassword()
		if user.PasswordHash != dbuser.PasswordHash {
			return NewError(2 /*Unauthorized Access*/)
		}

		err = DeleteSessionIfExists(context.Tx, r)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		sessionwriter, err := NewSession(context.Tx, r, dbuser.UserId)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
		return sessionwriter
	} else if r.Method == "GET" {
		s, err := GetSession(context.Tx, r)
		if err != nil {
			return NewError(1 /*Not Signed In*/)
		}

		return s
	} else if r.Method == "DELETE" {
		err := DeleteSessionIfExists(context.Tx, r)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
		return SuccessWriter{}
	}
	return NewError(3 /*Invalid Request*/)
}

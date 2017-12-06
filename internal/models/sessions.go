package models

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
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

func (s *Session) Cookie(domain string) *http.Cookie {
	return &http.Cookie{
		Name:     "moneygo-session",
		Value:    s.SessionSecret,
		Path:     "/",
		Domain:   domain,
		Expires:  s.Expires,
		Secure:   true,
		HttpOnly: true,
	}
}

func (s *Session) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func (s *Session) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(s)
}

func newSessionSecret() (string, error) {
	bits := make([]byte, 128)
	if _, err := io.ReadFull(rand.Reader, bits); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bits), nil
}

func NewSession(userid int64) (*Session, error) {
	session_secret, err := newSessionSecret()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	s := Session{
		SessionSecret: session_secret,
		UserId:        userid,
		Created:       now,
		Expires:       now.AddDate(0, 1, 0), // a month from now
	}

	return &s, nil
}

package handlers

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"log"
	"net/http"
	"time"
)

func GetSession(tx *Tx, r *http.Request) (*models.Session, error) {
	var s models.Session

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

type NewSessionWriter struct {
	session *models.Session
	cookie  *http.Cookie
}

func (n *NewSessionWriter) Write(w http.ResponseWriter) error {
	http.SetCookie(w, n.cookie)
	return n.session.Write(w)
}

func NewSession(tx *Tx, r *http.Request, userid int64) (*NewSessionWriter, error) {
	s, err := models.NewSession(userid)
	if err != nil {
		return nil, err
	}

	existing, err := tx.SelectInt("SELECT count(*) from sessions where SessionSecret=?", s.SessionSecret)
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, fmt.Errorf("%d session(s) exist with the generated session_secret", existing)
	}

	err = tx.Insert(s)
	if err != nil {
		return nil, err
	}

	return &NewSessionWriter{s, s.Cookie(r.URL.Host)}, nil
}

func SessionHandler(r *http.Request, context *Context) ResponseWriterWriter {
	if r.Method == "POST" || r.Method == "PUT" {
		var user models.User
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

package handlers

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"log"
	"net/http"
	"time"
)

func GetSession(tx store.Tx, r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie("moneygo-session")
	if err != nil {
		return nil, fmt.Errorf("moneygo-session cookie not set")
	}

	s, err := tx.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	if s.Expires.Before(time.Now()) {
		err := tx.DeleteSession(s)
		if err != nil {
			log.Printf("Unexpected error when attempting to delete expired session: %s", err)
		}
		return nil, fmt.Errorf("Session has expired")
	}
	return s, nil
}

func DeleteSessionIfExists(tx store.Tx, r *http.Request) error {
	session, err := GetSession(tx, r)
	if err == nil {
		err := tx.DeleteSession(session)
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

func NewSession(tx store.Tx, r *http.Request, userid int64) (*NewSessionWriter, error) {
	err := DeleteSessionIfExists(tx, r)
	if err != nil {
		return nil, err
	}

	s, err := models.NewSession(userid)
	if err != nil {
		return nil, err
	}

	exists, err := tx.SessionExists(s.SessionSecret)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("Session already exists with the generated session_secret")
	}

	err = tx.InsertSession(s)
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

		// Hash password before checking username to help mitigate timing
		// attacks
		user.HashPassword()

		dbuser, err := context.Tx.GetUserByUsername(user.Username)
		if err != nil {
			return NewError(2 /*Unauthorized Access*/)
		}

		if user.PasswordHash != dbuser.PasswordHash {
			return NewError(2 /*Unauthorized Access*/)
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

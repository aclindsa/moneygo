package db

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"time"
)

func (tx *Tx) InsertSession(session *models.Session) error {
	return tx.Insert(session)
}

func (tx *Tx) GetSession(secret string) (*models.Session, error) {
	var s models.Session

	err := tx.SelectOne(&s, "SELECT * from sessions where SessionSecret=?", secret)
	if err != nil {
		return nil, err
	}

	if s.Expires.Before(time.Now()) {
		tx.Delete(&s)
		return nil, fmt.Errorf("Session has expired")
	}
	return &s, nil
}

func (tx *Tx) SessionExists(secret string) (bool, error) {
	existing, err := tx.SelectInt("SELECT count(*) from sessions where SessionSecret=?", secret)
	return existing != 0, err
}

func (tx *Tx) DeleteSession(session *models.Session) error {
	_, err := tx.Delete(session)
	return err
}

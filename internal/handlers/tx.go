package handlers

import (
	"database/sql"
	"github.com/aclindsa/gorp"
	"strings"
)

type Tx struct {
	Dialect gorp.Dialect
	Tx      *gorp.Transaction
}

func (tx *Tx) Rebind(query string) string {
	chunks := strings.Split(query, "?")
	str := chunks[0]
	for i := 1; i < len(chunks); i++ {
		str += tx.Dialect.BindVar(i-1) + chunks[i]
	}
	return str
}

func (tx *Tx) Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	return tx.Tx.Select(i, tx.Rebind(query), args...)
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(tx.Rebind(query), args...)
}

func (tx *Tx) SelectInt(query string, args ...interface{}) (int64, error) {
	return tx.Tx.SelectInt(tx.Rebind(query), args...)
}

func (tx *Tx) SelectOne(holder interface{}, query string, args ...interface{}) error {
	return tx.Tx.SelectOne(holder, tx.Rebind(query), args...)
}

func (tx *Tx) Insert(list ...interface{}) error {
	return tx.Tx.Insert(list...)
}

func (tx *Tx) Update(list ...interface{}) (int64, error) {
	return tx.Tx.Update(list...)
}

func (tx *Tx) Delete(list ...interface{}) (int64, error) {
	return tx.Tx.Delete(list...)
}

func (tx *Tx) Commit() error {
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.Tx.Rollback()
}

func GetTx(db *gorp.DbMap) (*Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{db.Dialect, tx}, nil
}

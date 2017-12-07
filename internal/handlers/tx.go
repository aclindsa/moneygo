package handlers

import (
	"github.com/aclindsa/gorp"
	"github.com/aclindsa/moneygo/internal/store/db"
)

func GetTx(gdb *gorp.DbMap) (*db.Tx, error) {
	tx, err := gdb.Begin()
	if err != nil {
		return nil, err
	}
	return &db.Tx{gdb.Dialect, tx}, nil
}

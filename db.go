package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
	"log"
)

var DB *gorp.DbMap

func initDB(cfg *Config) {
	db, err := sql.Open(cfg.MoneyGo.DBType.String(), cfg.MoneyGo.DSN)
	if err != nil {
		log.Fatal(err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "UserId")
	dbmap.AddTableWithName(Session{}, "sessions").SetKeys(true, "SessionId")
	dbmap.AddTableWithName(Account{}, "accounts").SetKeys(true, "AccountId")
	dbmap.AddTableWithName(Security{}, "securities").SetKeys(true, "SecurityId")
	dbmap.AddTableWithName(Transaction{}, "transactions").SetKeys(true, "TransactionId")
	dbmap.AddTableWithName(Split{}, "splits").SetKeys(true, "SplitId")
	dbmap.AddTableWithName(Price{}, "prices").SetKeys(true, "PriceId")
	dbmap.AddTableWithName(Report{}, "reports").SetKeys(true, "ReportId")

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		log.Fatal(err)
	}

	DB = dbmap
}

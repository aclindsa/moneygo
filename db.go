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

	var dialect gorp.Dialect
	if cfg.MoneyGo.DBType == SQLite {
		dialect = gorp.SqliteDialect{}
	} else if cfg.MoneyGo.DBType == MySQL {
		dialect = gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		}
	} else if cfg.MoneyGo.DBType == Postgres {
		dialect = gorp.PostgresDialect{}
	} else {
		log.Fatalf("Don't know gorp dialect to go with '%s' DB type", cfg.MoneyGo.DBType.String())
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
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

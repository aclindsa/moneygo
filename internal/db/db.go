package db

import (
	"database/sql"
	"fmt"
	"github.com/aclindsa/moneygo/internal/config"
	"github.com/aclindsa/moneygo/internal/handlers"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
)

func GetDbMap(db *sql.DB, cfg *config.Config) (*gorp.DbMap, error) {
	var dialect gorp.Dialect
	if cfg.MoneyGo.DBType == config.SQLite {
		dialect = gorp.SqliteDialect{}
	} else if cfg.MoneyGo.DBType == config.MySQL {
		dialect = gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		}
	} else if cfg.MoneyGo.DBType == config.Postgres {
		dialect = gorp.PostgresDialect{}
	} else {
		return nil, fmt.Errorf("Don't know gorp dialect to go with '%s' DB type", cfg.MoneyGo.DBType.String())
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	dbmap.AddTableWithName(handlers.User{}, "users").SetKeys(true, "UserId")
	dbmap.AddTableWithName(handlers.Session{}, "sessions").SetKeys(true, "SessionId")
	dbmap.AddTableWithName(handlers.Account{}, "accounts").SetKeys(true, "AccountId")
	dbmap.AddTableWithName(handlers.Security{}, "securities").SetKeys(true, "SecurityId")
	dbmap.AddTableWithName(handlers.Transaction{}, "transactions").SetKeys(true, "TransactionId")
	dbmap.AddTableWithName(handlers.Split{}, "splits").SetKeys(true, "SplitId")
	dbmap.AddTableWithName(handlers.Price{}, "prices").SetKeys(true, "PriceId")
	dbmap.AddTableWithName(handlers.Report{}, "reports").SetKeys(true, "ReportId")

	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}

	return dbmap, nil
}

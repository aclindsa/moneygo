package db

import (
	"database/sql"
	"fmt"
	"github.com/aclindsa/gorp"
	"github.com/aclindsa/moneygo/internal/config"
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

const luaMaxLengthBuffer int = 4096

func GetDbMap(db *sql.DB, dbtype config.DbType) (*gorp.DbMap, error) {
	var dialect gorp.Dialect
	if dbtype == config.SQLite {
		dialect = gorp.SqliteDialect{}
	} else if dbtype == config.MySQL {
		dialect = gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		}
	} else if dbtype == config.Postgres {
		dialect = gorp.PostgresDialect{
			LowercaseFields: true,
		}
	} else {
		return nil, fmt.Errorf("Don't know gorp dialect to go with '%s' DB type", dbtype.String())
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	dbmap.AddTableWithName(models.User{}, "users").SetKeys(true, "UserId")
	dbmap.AddTableWithName(handlers.Session{}, "sessions").SetKeys(true, "SessionId")
	dbmap.AddTableWithName(handlers.Account{}, "accounts").SetKeys(true, "AccountId")
	dbmap.AddTableWithName(handlers.Security{}, "securities").SetKeys(true, "SecurityId")
	dbmap.AddTableWithName(handlers.Transaction{}, "transactions").SetKeys(true, "TransactionId")
	dbmap.AddTableWithName(handlers.Split{}, "splits").SetKeys(true, "SplitId")
	dbmap.AddTableWithName(handlers.Price{}, "prices").SetKeys(true, "PriceId")
	rtable := dbmap.AddTableWithName(handlers.Report{}, "reports").SetKeys(true, "ReportId")
	rtable.ColMap("Lua").SetMaxSize(handlers.LuaMaxLength + luaMaxLengthBuffer)

	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}

	return dbmap, nil
}

func GetDSN(dbtype config.DbType, dsn string) string {
	if dbtype == config.MySQL && !strings.Contains(dsn, "parseTime=true") {
		log.Fatalf("The DSN for MySQL MUST contain 'parseTime=True' but does not!")
	}
	return dsn
}

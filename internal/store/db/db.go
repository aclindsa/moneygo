package db

import (
	"database/sql"
	"fmt"
	"github.com/aclindsa/gorp"
	"github.com/aclindsa/moneygo/internal/config"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

// luaMaxLengthBuffer is intended to be enough bytes such that a given string
// no longer than models.LuaMaxLength is sure to fit within a database
// implementation's string type specified by the same.
const luaMaxLengthBuffer int = 4096

func getDbMap(db *sql.DB, dbtype config.DbType) (*gorp.DbMap, error) {
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
	dbmap.AddTableWithName(models.Session{}, "sessions").SetKeys(true, "SessionId")
	dbmap.AddTableWithName(models.Security{}, "securities").SetKeys(true, "SecurityId")
	dbmap.AddTableWithName(Price{}, "prices").SetKeys(true, "PriceId")
	dbmap.AddTableWithName(models.Account{}, "accounts").SetKeys(true, "AccountId")
	dbmap.AddTableWithName(models.Transaction{}, "transactions").SetKeys(true, "TransactionId")
	dbmap.AddTableWithName(Split{}, "splits").SetKeys(true, "SplitId")
	rtable := dbmap.AddTableWithName(models.Report{}, "reports").SetKeys(true, "ReportId")
	rtable.ColMap("Lua").SetMaxSize(models.LuaMaxLength + luaMaxLengthBuffer)

	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}

	return dbmap, nil
}

func getDSN(dbtype config.DbType, dsn string) string {
	if dbtype == config.MySQL && !strings.Contains(dsn, "parseTime=true") {
		log.Fatalf("The DSN for MySQL MUST contain 'parseTime=True' but does not!")
	}
	return dsn
}

type DbStore struct {
	dbMap *gorp.DbMap
}

func (db *DbStore) Empty() error {
	return db.dbMap.TruncateTables()
}

func (db *DbStore) Begin() (store.Tx, error) {
	tx, err := db.dbMap.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{db.dbMap.Dialect, tx}, nil
}

func (db *DbStore) Close() error {
	err := db.dbMap.Db.Close()
	db.dbMap = nil
	return err
}

func GetStore(dbtype config.DbType, dsn string) (store store.Store, err error) {
	dsn = getDSN(dbtype, dsn)
	database, err := sql.Open(dbtype.String(), dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			database.Close()
		}
	}()

	dbmap, err := getDbMap(database, dbtype)
	if err != nil {
		return nil, err
	}
	return &DbStore{dbmap}, nil
}

package config

import (
	"errors"
	"fmt"
	"gopkg.in/gcfg.v1"
	"strings"
)

type DbType uint

const (
	SQLite DbType = 1 + iota
	MySQL
	Postgres
)

var dbTypes = [...]string{"sqlite3", "mysql", "postgres"}

func (e DbType) Valid() bool {
	// This check is mostly out of paranoia, ensuring e != 0 should be
	// sufficient
	return e >= SQLite && e <= Postgres
}

func (e DbType) String() string {
	if e.Valid() {
		return dbTypes[e-1]
	}
	return fmt.Sprintf("invalid DbType (%d)", e)
}

func (e *DbType) FromString(in string) error {
	value := strings.TrimSpace(in)

	for i, s := range dbTypes {
		if s == value {
			*e = DbType(i + 1)
			return nil
		}
	}
	*e = 0
	return errors.New("Invalid DbType: \"" + in + "\"")
}

func (e *DbType) UnmarshalText(text []byte) error {
	return e.FromString(string(text))
}

type MoneyGo struct {
	Fcgi    bool   // whether to serve FCGI (HTTP by default if false)
	Port    int    // port to serve API/files on
	Basedir string `gcfg:"base-directory"` // base directory for serving files out of
	DBType  DbType `gcfg:"db-type"`        // Whether this is a sqlite/mysql/postgresql database
	DSN     string `gcfg:"db-dsn"`         // 'Data Source Name' for database connection
}

type Https struct {
	CertFile           string `gcfg:"cert-file"`
	KeyFile            string `gcfg:"key-file"`
	GenerateCerts      bool   `gcfg:"generate-certs-if-absent"` // Generate certificates if missing
	GenerateCertsHosts string `gcfg:"generate-certs-hosts"`     // Hostnames to generate certificates for if missing and GenerateCerts==true
}

type Config struct {
	MoneyGo MoneyGo
	Https   Https
}

func ReadConfig(filename string) (*Config, error) {
	cfg := Config{
		MoneyGo: MoneyGo{
			Fcgi:    false,
			Port:    80,
			Basedir: "src/github.com/aclindsa/moneygo/",
			DBType:  SQLite,
			DSN:     "file:moneygo.sqlite?cache=shared&mode=rwc",
		},
		Https: Https{
			CertFile:           "./cert.pem",
			KeyFile:            "./key.pem",
			GenerateCerts:      false,
			GenerateCertsHosts: "localhost",
		},
	}

	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %s", err)
	}
	return &cfg, nil
}

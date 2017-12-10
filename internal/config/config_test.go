package config_test

import (
	"github.com/aclindsa/moneygo/internal/config"
	"testing"
)

func TestSqliteHTTPSConfig(t *testing.T) {
	cfg, err := config.ReadConfig("./testdata/sqlite_https_config.ini")
	if err != nil {
		t.Fatalf("Unexpected error parsing config: %s\n", err)
	}

	if cfg.MoneyGo.Fcgi {
		t.Errorf("MoneyGo.Fcgi unexpectedly true")
	}
	if cfg.MoneyGo.Port != 8443 {
		t.Errorf("MoneyGo.Port %d instead of 8443", cfg.MoneyGo.Port)
	}
	if cfg.MoneyGo.Basedir != "src/github.com/aclindsa/moneygo/" {
		t.Errorf("MoneyGo.Basedir not correct")
	}
	if cfg.MoneyGo.DBType != config.SQLite {
		t.Errorf("MoneyGo.DBType not config.SQLite")
	}
	if cfg.MoneyGo.DSN != "file:moneygo.sqlite?cache=shared&mode=rwc" {
		t.Errorf("MoneyGo.DSN not correct")
	}

	if cfg.Https.CertFile != "./cert.pem" {
		t.Errorf("Https.CertFile '%s', not ./cert.pem", cfg.Https.CertFile)
	}
	if cfg.Https.KeyFile != "./key.pem" {
		t.Errorf("Https.KeyFile '%s', not ./key.pem", cfg.Https.KeyFile)
	}
	if cfg.Https.GenerateCerts {
		t.Errorf("Https.GenerateCerts not false")
	}
	if cfg.Https.GenerateCertsHosts != "localhost,127.0.0.1" {
		t.Errorf("Https.GenerateCertsHosts '%s', not localhost", cfg.Https.GenerateCertsHosts)
	}
}

func TestPostgresFcgiConfig(t *testing.T) {
	cfg, err := config.ReadConfig("./testdata/postgres_fcgi_config.ini")
	if err != nil {
		t.Fatalf("Unexpected error parsing config: %s\n", err)
	}

	if !cfg.MoneyGo.Fcgi {
		t.Errorf("MoneyGo.Fcgi unexpectedly false")
	}
	if cfg.MoneyGo.Port != 9001 {
		t.Errorf("MoneyGo.Port %d instead of 9001", cfg.MoneyGo.Port)
	}
	if cfg.MoneyGo.Basedir != "src/github.com/aclindsa/moneygo/" {
		t.Errorf("MoneyGo.Basedir not correct")
	}
	if cfg.MoneyGo.DBType != config.Postgres {
		t.Errorf("MoneyGo.DBType not config.Postgres")
	}
	if cfg.MoneyGo.DSN != "postgres://moneygo_test@localhost/moneygo_test?sslmode=disable" {
		t.Errorf("MoneyGo.DSN not correct")
	}
}

func TestGenerateCertsConfig(t *testing.T) {
	cfg, err := config.ReadConfig("./testdata/generate_certs_config.ini")
	if err != nil {
		t.Fatalf("Unexpected error parsing config: %s\n", err)
	}

	if cfg.Https.CertFile != "./local_cert.pem" {
		t.Errorf("Https.CertFile '%s', not ./local_cert.pem", cfg.Https.CertFile)
	}
	if cfg.Https.KeyFile != "./local_key.pem" {
		t.Errorf("Https.KeyFile '%s', not ./local_key.pem", cfg.Https.KeyFile)
	}
	if !cfg.Https.GenerateCerts {
		t.Errorf("Https.GenerateCerts not true")
	}
	if cfg.Https.GenerateCertsHosts != "example.com" {
		t.Errorf("Https.GenerateCertsHosts '%s', not example.com", cfg.Https.GenerateCertsHosts)
	}
}

func TestNonexistentConfig(t *testing.T) {
	cfg, err := config.ReadConfig("./testdata/nonexistent_config.ini")
	if err == nil || cfg != nil {
		t.Fatalf("Expected error parsing nonexistent config")
	}
}

package db

import (
	"testing"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/sirupsen/logrus"
)

func TestInitDB_NoDSN(t *testing.T) {
	cfg := config.Config{
		DatabaseDsn: "",
	}
	logger := logrus.New()
	db, err := InitDB(cfg, logger)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db == nil {
		t.Fatal("expected db instance")
	}
}

func TestInitDB_WithDSN(t *testing.T) {
	cfg := config.Config{
		DatabaseDsn: "postgres://invalid",
	}

	logger := logrus.New()
	db, err := InitDB(cfg, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db == nil {
		t.Fatal("expected db instance")
	}
}
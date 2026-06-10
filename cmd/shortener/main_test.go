package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
)

func TestSelectStorage_DB(t *testing.T) {
	cfg := config.Config{
		DatabaseDsn:     "postgres://prikol",
		FileStoragePath: "tmp",
	}
	db := &sql.DB{}

	storage, err := selectStorage(cfg, db)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage == nil {
		t.Fatal("expected storage")
	}
	os.Remove("tmp")
}

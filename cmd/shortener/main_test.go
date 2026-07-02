package main

import (
	"bytes"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func TestSelectStorage_File(t *testing.T) {
	dir := t.TempDir()

	cfg := config.Config{
		FileStoragePath: filepath.Join(dir, "storage.json"),
	}

	storage, err := selectStorage(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage == nil {
		t.Fatal("expected storage")
	}

	defer storage.Close()
}

func TestSelectStorage_OpenError(t *testing.T) {
	cfg := config.Config{
		FileStoragePath: "/not/existing/path/storage.json",
	}

	storage, err := selectStorage(cfg, nil)

	if err == nil {
		t.Fatal("expected error")
	}
	if storage != nil {
		t.Fatal("expected nil storage")
	}
}

func TestPrintBuildInfo(t *testing.T) {
	oldVersion := buildVersion
	oldDate := buildDate
	oldCommit := buildCommit
	oldWriter := log.Writer()
	oldFlags := log.Flags()

	defer func() {
		buildVersion = oldVersion
		buildDate = oldDate
		buildCommit = oldCommit
		log.SetOutput(oldWriter)
		log.SetFlags(oldFlags)
	}()

	buildVersion = "1.2.3"
	buildDate = "2025-01-01"
	buildCommit = "abcdef"

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)

	printBuildInfo()

	got := buf.String()

	want := []string{
		"Build version: 1.2.3",
		"Build date: 2025-01-01",
		"Build commit: abcdef",
	}

	for _, s := range want {
		if !strings.Contains(got, s) {
			t.Errorf("log does not contain %q\nactual:\n%s", s, got)
		}
	}
}

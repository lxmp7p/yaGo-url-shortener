package repository

import (
	"database/sql"
	"testing"
)

func TestNewDatabaseCache(t *testing.T) {
	db := &sql.DB{}

	cache := NewDatabaseCache(db)
	if cache == nil {
		t.Fatal("expected cache, got nil")
	}
	if cache.db != db {
		t.Fatal("db not assigned correctly")
	}
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestDatabaseCache_Save_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	mock.ExpectExec("INSERT INTO urls").
		WithArgs(sqlmock.AnyArg(), "https://google.com", "abc", "user1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cache.Save(context.Background(), "https://google.com", "abc", "user1")
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseCache_Save_AlreadyExists(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	mock.ExpectExec("INSERT INTO urls").
		WithArgs(sqlmock.AnyArg(), "https://google.com", "abc", "user1").
		WillReturnResult(sqlmock.NewResult(1, 0))

	rows := sqlmock.NewRows([]string{"short"}).
		AddRow("existing")

	mock.ExpectQuery("SELECT short FROM urls").
		WithArgs("https://google.com").
		WillReturnRows(rows)

	err := cache.Save(context.Background(), "https://google.com", "abc", "user1")

	if _, ok := err.(*URLError); !ok {
		t.Fatalf("expected URLError got %T", err)
	}
}

func TestDatabaseCache_Get_OK(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	rows := sqlmock.NewRows([]string{"original", "is_deleted"}).
		AddRow("https://google.com", false)

	mock.ExpectQuery("SELECT original").
		WithArgs("abc").
		WillReturnRows(rows)

	url, err := cache.Get(context.Background(), "abc")

	if err != nil {
		t.Fatal(err)
	}

	if url != "https://google.com" {
		t.Fatal("wrong url")
	}
}

func TestDatabaseCache_Get_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	mock.ExpectQuery("SELECT original").
		WithArgs("abc").
		WillReturnError(sql.ErrNoRows)

	_, err := cache.Get(context.Background(), "abc")

	if !errors.Is(err, ErrURLNotFound) {
		t.Fatal(err)
	}
}

func TestDatabaseCache_Get_Deleted(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	rows := sqlmock.NewRows([]string{"original", "is_deleted"}).
		AddRow("https://google.com", true)

	mock.ExpectQuery("SELECT original").
		WithArgs("abc").
		WillReturnRows(rows)

	_, err := cache.Get(context.Background(), "abc")

	if !errors.Is(err, ErrURLDeleted) {
		t.Fatal(err)
	}
}

func TestDatabaseCache_GetByUserID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	rows := sqlmock.NewRows([]string{"id", "original", "short"}).
		AddRow("550e8400-e29b-41d4-a716-446655440000", "https://google.com", "abc").
		AddRow("6ba7b810-9dad-11d1-80b4-00c04fd430c8", "https://ya.ru", "def")

	mock.ExpectQuery("SELECT id, original, short").
		WithArgs("user1").
		WillReturnRows(rows)

	urls, err := cache.GetByUserID(context.Background(), "user1")

	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != 2 {
		t.Fatal("expected 2 urls")
	}
}

func TestDatabaseCache_Load(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cache := NewDatabaseCache(db)

	rows := sqlmock.NewRows([]string{"id", "original", "short"}).
		AddRow("6ba7b810-9dad-11d1-80b4-00c04fd430c8", "https://google.com", "abc")

	mock.ExpectQuery("SELECT id, original, short FROM urls").
		WillReturnRows(rows)

	urls, err := cache.Load(context.Background(), "", "")

	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != 1 {
		t.Fatal("expected 1 url")
	}
}

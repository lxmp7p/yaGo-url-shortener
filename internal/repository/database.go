package repository

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/google/uuid"
)

type DatabaseCache struct {
	mu sync.RWMutex
	db *sql.DB
}

type URL struct {
	uuid     uuid.UUID
	original string
	short    string
}

func NewDatabaseCache(db *sql.DB) *DatabaseCache {
	cache := &DatabaseCache{
		db: db,
	}

	return cache
}

func (cache *DatabaseCache) Save(originalURL, shortURL string) error {
	if _, exists := cache.Get(shortURL); exists == nil {
		return ErrExist
	}

	query := "INSERT INTO urls (id, original, short) VALUES ($1, $2, $3)"
	_, err := cache.db.Exec(query, uuid.New(), originalURL, shortURL)
	if err != nil {
		return err
	}

	return nil
}

func (cache *DatabaseCache) Get(shortURL string) (string, error) {
	query := "SELECT original FROM urls WHERE short = $1"

	var original string
	err := cache.db.QueryRow(query, shortURL).Scan(&original)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New(urlNotFound)
		}
		return "", err
	}
	return original, nil
}

func (cache *DatabaseCache) Load(shortURL string) ([]*URL, error) {
	query := "SELECT id, original, short FROM urls"

	rows, err := cache.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*URL
	for rows.Next() {
		var u URL
		if err = rows.Scan(&u.uuid, &u.original, &u.short); err != nil {
			return nil, err
		}
		result = append(result, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

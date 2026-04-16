package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type DatabaseCache struct {
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

type URLError struct {
	Short string
}

func (e *URLError) Error() string {
	return ErrOriginalURLExists.Error()
}

func (cache *DatabaseCache) Save(ctx context.Context, originalURL, shortURL string, userID string) error {
	query := `
	INSERT INTO urls (id, original, short) 
	VALUES ($1, $2, $3) 
	ON CONFLICT (original) DO NOTHING 
	`
	result, err := cache.db.Exec(query, uuid.New(), originalURL, shortURL)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		var existing string
		err := cache.db.QueryRow(
			`SELECT short FROM urls WHERE original = $1`,
			originalURL,
		).Scan(&existing)

		if err != nil {
			return err
		}

		return &URLError{Short: existing}
	}

	return nil
}

func (cache *DatabaseCache) Get(ctx context.Context, shortURL string) (string, error) {
	query := "SELECT original FROM urls WHERE short = $1"

	var original string
	err := cache.db.QueryRow(query, shortURL).Scan(&original)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrURLNotFound
		}
		return "", err
	}
	return original, nil
}

func (cache *DatabaseCache) Load(ctx context.Context, shortURL string, userID string) ([]*URL, error) {
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

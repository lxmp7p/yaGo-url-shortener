package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// generate:reset
type DatabaseCache struct {
	db *sql.DB
}

func NewDatabaseCache(db *sql.DB) *DatabaseCache {
	cache := &DatabaseCache{
		db: db,
	}

	return cache
}

func (cache *DatabaseCache) Close() error {
	return cache.db.Close()
}

// generate:reset
type URLError struct {
	Short string
}

func (e *URLError) Error() string {
	return ErrOriginalURLExists.Error()
}

// Возвращает количество сохранённых URL и уникальных пользователей
func (cache *DatabaseCache) Stats(ctx context.Context) (int, int, error) {
	var urls int
	if err := cache.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM urls").Scan(&urls); err != nil {
		return 0, 0, err
	}

	var users int
	query := "SELECT COUNT(DISTINCT owner_id) FROM urls"
	if err := cache.db.QueryRowContext(ctx, query).Scan(&users); err != nil {
		return 0, 0, err
	}

	return urls, users, nil
}

func (cache *DatabaseCache) Save(ctx context.Context, originalURL, shortURL string, userID string) error {
	query := `
	INSERT INTO urls (id, original, short, owner_id) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (original) DO NOTHING 
	`
	result, err := cache.db.Exec(query, uuid.New(), originalURL, shortURL, userID)
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
	query := "SELECT original, is_deleted FROM urls WHERE short = $1"

	var original string
	var isDeleted bool

	err := cache.db.QueryRowContext(ctx, query, shortURL).
		Scan(&original, &isDeleted)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrURLNotFound
		}
		return "", err
	}

	if isDeleted {
		return "", ErrURLDeleted
	}
	return original, nil
}

func (cache *DatabaseCache) GetByUserID(ctx context.Context, ownerID string) ([]URL, error) {
	query := "SELECT id, original, short FROM urls WHERE owner_id = $1"

	rows, err := cache.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []URL
	for rows.Next() {
		var u URL
		if err = rows.Scan(&u.UUID, &u.Original, &u.Short); err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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
		if err = rows.Scan(&u.UUID, &u.Original, &u.Short); err != nil {
			return nil, err
		}
		result = append(result, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (cache *DatabaseCache) Delete(ctx context.Context, userID string, IDs []string) error {
	query := "UPDATE urls SET is_deleted = TRUE WHERE owner_id = $1 AND short = ANY($2)"

	_, err := cache.db.ExecContext(ctx, query, userID, IDs)
	if err != nil {
		return err
	}

	return err
}

package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrExist = errors.New("shortURL already exists")
)

type Cache struct {
	URLCache map[string]URL
	mu       sync.RWMutex
	filename string
}

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

type URL struct {
	uuid     uuid.UUID
	Original string `json:"short_url"`
	Short    string `json:"original_url"`
	UserID   string `json:"user_id"`
}

func NewCache(ctx context.Context, filepath string) *Cache {
	cache := &Cache{
		URLCache: make(map[string]URL),
		filename: filepath,
	}

	cache.Load(ctx, filepath)

	return cache
}

func (cache *Cache) Save(ctx context.Context, originalURL, shortURL, userID string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, exists := cache.URLCache[shortURL]; exists {
		return ErrExist
	}

	cache.URLCache[shortURL] = URL{
		Original: originalURL,
		Short:    shortURL,
		UserID:   userID,
	}

	record := Record{
		UUID:        uuid.NewString(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(cache.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return nil
}

func (cache *Cache) Get(ctx context.Context, shortURL string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	URL, ok := cache.URLCache[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	return URL.Original, nil
}

func (cache *Cache) GetByUserID(ctx context.Context, userID string) ([]URL, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	var urls []URL
	for _, v := range cache.URLCache {
		if v.UserID == userID {
			urls = append(urls, v)
		}
	}

	return urls, nil
}

func (cache *Cache) Load(ctx context.Context, shortURL string) (*Cache, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	file, err := os.OpenFile(cache.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return &Cache{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		record := Record{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}
		cache.URLCache[record.ShortURL] = URL{Original: record.OriginalURL}
	}
	return cache, nil
}

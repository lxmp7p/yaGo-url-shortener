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
	URLCache map[string]string
	mu       sync.RWMutex
	filename string
}

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewCache(ctx context.Context, filepath string) *Cache {
	cache := &Cache{
		URLCache: make(map[string]string),
		filename: filepath,
	}

	cache.Load(ctx, filepath)

	return cache
}

func (cache *Cache) Save(ctx context.Context, originalURL, shortURL string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, exists := cache.URLCache[shortURL]; exists {
		return ErrExist
	}
	cache.URLCache[shortURL] = originalURL

	record := Record{
		UUID:        uuid.NewString(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
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

	originalURL, ok := cache.URLCache[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	return originalURL, nil
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
		cache.URLCache[record.ShortURL] = record.OriginalURL
	}
	return cache, nil
}

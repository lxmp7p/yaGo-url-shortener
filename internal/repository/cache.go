package repository

import (
	"errors"
	"sync"
)

var (
	ErrExist = errors.New("shortURL already exists")
)

type Cache struct {
	URLCache map[string]string
	mu       sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		URLCache: make(map[string]string),
	}
}

func (cache *Cache) Save(originalURL, shortURL string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, exists := cache.URLCache[shortURL]; exists {
		return ErrExist
	}
	cache.URLCache[shortURL] = originalURL

	return nil
}

func (cache *Cache) Get(shortURL string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	originalURL, ok := cache.URLCache[shortURL]
	if !ok {
		return "", errors.New("url not found")
	}

	return originalURL, nil
}

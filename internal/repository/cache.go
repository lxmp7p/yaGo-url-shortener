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

// Структура реализации кэша
// generate:reset
type Cache struct {
	URLCache map[string]URL
	mu       sync.RWMutex
	filename string
	file     *os.File
}

// Структура для хранения информации о записи
// generate:reset
type URL struct {
	UUID        uuid.UUID `json:"id"`
	Original    string    `json:"original_url"`
	Short       string    `json:"short_url"`
	UserID      string    `json:"user_id"`
	DeletedFlag bool      `json:"is_deleted"`
}

// Создает новую структуру Cache
func NewCache(ctx context.Context, filepath string, file *os.File) *Cache {
	cache := &Cache{
		URLCache: make(map[string]URL),
		filename: filepath,
		file:     file,
	}

	cache.Load(ctx, filepath)

	return cache
}

func (cache *Cache) Close() error {
	return cache.file.Close()
}

// Проверяет наличие shortURL в кэше и в случае отсутствия, добавляет
// новое значение URL{} в кэш
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

	data, err := json.Marshal(URL{
		UUID:     uuid.New(),
		Short:    shortURL,
		Original: originalURL,
		UserID:   userID,
	})

	if err != nil {
		return err
	}

	_, err = cache.file.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return nil
}

// Возвращает значение из кэша. В случае, если такого значения нет
// возвращает ошибку ErrURLNotFound
func (cache *Cache) Get(ctx context.Context, shortURL string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	URL, ok := cache.URLCache[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	return URL.Original, nil
}

// Возвращает все значения сохраненные пользователем по его userID
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

// Открывает файл на чтение, получает из него все сохраненные записи
// и записывает их в Cache
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
		line := scanner.Bytes()
		record := URL{}
		if err := json.Unmarshal(line, &record); err != nil {
			continue
		}
		cache.URLCache[record.Short] = record
	}
	return cache, nil
}

func (cache *Cache) Delete(ctx context.Context, shortURL string, IDs []string) error {
	return nil
}

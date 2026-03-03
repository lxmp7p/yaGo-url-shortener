package repository

type Cache struct {
	URLCache map[string]string
}

func NewCache() *Cache {
	return &Cache{
		URLCache: make(map[string]string),
	}
}

func (cache *Cache) Save(originalURL, shortURL string) {
	cache.URLCache[shortURL] = originalURL
}

func (cache *Cache) Get(shortURL string) (string, bool) {
	originalURL, ok := cache.URLCache[shortURL]
	return originalURL, ok
}

func (cache *Cache) Exist(shortURL string) bool {
	_, ok := cache.URLCache[shortURL]
	return ok
}

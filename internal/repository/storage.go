package repository

type URLstorage interface {
	Save(originalURL string, shortURL string)
	Get(shortURL string) (string, bool)
	Exist(shortURL string) bool
}

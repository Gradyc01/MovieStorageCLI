package actions

import (
	"movie-tracker/internal/storage"
)

type Store struct {
	storage storage.Storage
}

func CreateStore(filePath string) *Store {
	return &Store{
		storage: storage.NewJSONStorage(filePath),
	}
}

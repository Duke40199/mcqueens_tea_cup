package infra

import (
	"encoding/json"
	"os"
	"sync"
)

type FileStore struct {
	filename string
	mutex    sync.Mutex
	data     map[string]string
}

func NewFileStore(filename string) *FileStore {
	store := &FileStore{
		filename: filename,
		data:     make(map[string]string),
	}
	store.load()
	return store
}

func (fs *FileStore) load() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	file, err := os.ReadFile(fs.filename)
	if err != nil {
		return // File doesn't exist yet, that's fine
	}
	json.Unmarshal(file, &fs.data)
}

func (fs *FileStore) GetLastSeen(feedURL string) (string, bool) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	val, ok := fs.data[feedURL]
	return val, ok
}

func (fs *FileStore) SetLastSeen(feedURL, guid string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.data[feedURL] = guid

	// Save immediately to disk
	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filename, data, 0644)
}

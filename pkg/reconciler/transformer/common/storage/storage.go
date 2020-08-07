package storage

import "sync"

// Storage is a simple object that provides thread safe
// methods to read and write into a map.
type Storage struct {
	data map[string]interface{}
	mux  sync.RWMutex
}

// New returns an instance of Storage.
func New() *Storage {
	return &Storage{
		data: make(map[string]interface{}),
		mux:  sync.RWMutex{},
	}
}

// Set writes a value interface to a string key.
func (s *Storage) Set(k string, v interface{}) {
	s.mux.Lock()
	s.data[k] = v
	s.mux.Unlock()
}

// Get reads value by a key.
func (s *Storage) Get(k string) interface{} {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.data[k]
}

// GetString reads value by a key and asserts String type.
func (s *Storage) GetString(k string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	str, ok := s.data[k].(string)
	return str, ok
}

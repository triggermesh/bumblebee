package storage

import "sync"

type Storage struct {
	data
	mux sync.RWMutex
}

type data map[string]interface{}

func New() *Storage {
	return &Storage{
		data: make(map[string]interface{}),
		mux:  sync.RWMutex{},
	}
}

func (s *Storage) Set(k string, v interface{}) {
	s.mux.Lock()
	s.data[k] = v
	s.mux.Unlock()
}

func (s *Storage) Get(k string) interface{} {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.data[k]
}

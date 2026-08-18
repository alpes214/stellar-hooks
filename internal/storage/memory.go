package storage

import "sync"

type MemoryCursorStore struct {
	mu      sync.Mutex
	cursors map[string]string
}

func NewMemoryCursorStore() *MemoryCursorStore {
	return &MemoryCursorStore{cursors: map[string]string{}}
}

func (s *MemoryCursorStore) GetCursor(stream string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cursors[stream], nil
}

func (s *MemoryCursorStore) SetCursor(stream string, cursor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cursors[stream] = cursor
	return nil
}

package safeset

import "sync"

type SafeSet struct {
	mu sync.Mutex
	v  map[string]struct{}
}

func NewSafeSet() *SafeSet {
	return &SafeSet{
		v: make(map[string]struct{}),
	}
}

func (s *SafeSet) Insert(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.v[key] = struct{}{}
}

func (s *SafeSet) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.v[key]
	if ok {
		delete(s.v, key)
	}
}

func (s *SafeSet) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.v[key]
	return ok
}

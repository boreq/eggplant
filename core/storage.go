package core

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		storage: make(map[string]*Data),
	}
}

type memoryStorage struct {
	storage map[string]*Data
}

func (s *memoryStorage) Get(key string) (*Data, bool) {
	data, ok := s.storage[key]
	return data, ok
}

func (s *memoryStorage) Put(key string, value *Data) {
	s.storage[key] = value
}

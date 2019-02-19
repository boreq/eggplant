package core

import (
	"strconv"
)

func NewSet() *Set {
	return &Set{
		entries: make(map[string]struct{}),
	}
}

type Set struct {
	entries map[string]struct{}
}

func (s *Set) Add(value string) {
	s.entries[value] = struct{}{}
}

func (s *Set) Size() int {
	return len(s.entries)
}

func (s *Set) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(s.Size())), nil
}

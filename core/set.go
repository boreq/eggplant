package core

import (
	"strconv"
)

func NewSet() Set {
	return nil
}

type Set []string

func (s *Set) Add(value string) {
	if !s.Contains(value) {
		*s = cheapAppend(*s, value)
	}
}

func (s Set) Contains(value string) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}
	return false
}

func (s Set) Size() int {
	return len(s)
}

func (s *Set) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(s.Size())), nil
}

func cheapAppend(slice []string, value string) []string {
	newSlice := make([]string, len(slice)+1, len(slice)+1)
	copy(newSlice, slice)
	newSlice[len(newSlice)-1] = value
	return newSlice
}

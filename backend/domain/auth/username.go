package auth

import (
	"fmt"

	"github.com/boreq/errors"
)

const maxUsernameLen = 100

type Username struct {
	value string
}

func NewUsernameFromString(s string) (Username, error) {
	if s == "" {
		return Username{}, errors.New("username can't be empty")
	}
	if len(s) > maxUsernameLen {
		return Username{}, fmt.Errorf("username length can't exceed %d characters", maxUsernameLen)
	}
	return Username{value: s}, nil
}

func (u Username) String() string {
	return u.value
}

package remotecontrol

import (
	"strings"
	"unicode/utf8"

	"github.com/boreq/errors"
)

const maxDeviceNameLength = 64

type DeviceName struct {
	value string
}

func NewDeviceName(s string) (DeviceName, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return DeviceName{}, errors.New("device name must not be empty")
	}
	if !utf8.ValidString(s) {
		return DeviceName{}, errors.New("device name must be valid utf-8")
	}
	if utf8.RuneCountInString(s) > maxDeviceNameLength {
		return DeviceName{}, errors.New("device name is too long")
	}
	return DeviceName{value: s}, nil
}

func (n DeviceName) String() string {
	return n.value
}

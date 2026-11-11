package remotecontrol

import (
	"github.com/boreq/errors"
	"github.com/oklog/ulid/v2"
)

const maxDeviceIDLength = 64

type DeviceID struct {
	value string
}

func NewDeviceID(s string) (DeviceID, error) {
	if s == "" {
		return DeviceID{}, errors.New("device id must not be empty")
	}
	if len(s) > maxDeviceIDLength {
		return DeviceID{}, errors.New("device id is too long")
	}
	for _, r := range s {
		if !isDeviceIDRune(r) {
			return DeviceID{}, errors.New("device id contains an invalid character")
		}
	}
	return DeviceID{value: s}, nil
}

func (id DeviceID) String() string {
	return id.value
}

func isDeviceIDRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
}

type ConnectionID struct {
	value ulid.ULID
}

func NewConnectionID() ConnectionID {
	return ConnectionID{value: ulid.Make()}
}

func (id ConnectionID) String() string {
	return id.value.String()
}

package remote

import (
	"github.com/boreq/errors"
	"github.com/oklog/ulid/v2"
)

type RemoteInstanceID struct {
	value ulid.ULID
}

func NewRemoteInstanceID() (RemoteInstanceID, error) {
	return RemoteInstanceID{value: ulid.Make()}, nil
}

func NewRemoteInstanceIDFromString(s string) (RemoteInstanceID, error) {
	if s == "" {
		return RemoteInstanceID{}, errors.New("remote instance id must not be empty")
	}
	id, err := ulid.Parse(s)
	if err != nil {
		return RemoteInstanceID{}, errors.Wrap(err, "could not parse the ulid")
	}
	return RemoteInstanceID{value: id}, nil
}

func (id RemoteInstanceID) String() string {
	return id.value.String()
}

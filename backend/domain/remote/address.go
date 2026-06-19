package remote

import (
	"net/url"
	"strings"

	"github.com/boreq/errors"
)

type RemoteInstanceAddress struct {
	value string
}

func NewRemoteInstanceAddress(s string) (RemoteInstanceAddress, error) {
	if s == "" {
		return RemoteInstanceAddress{}, errors.New("address must not be empty")
	}

	u, err := url.Parse(s)
	if err != nil {
		return RemoteInstanceAddress{}, errors.Wrap(err, "could not parse the address")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return RemoteInstanceAddress{}, errors.New("address scheme must be http or https")
	}

	if u.Host == "" {
		return RemoteInstanceAddress{}, errors.New("address must have a host")
	}

	return RemoteInstanceAddress{value: strings.TrimRight(s, "/")}, nil
}

func (a RemoteInstanceAddress) String() string {
	return a.value
}

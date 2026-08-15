package remote

import (
	"net/url"

	"github.com/boreq/errors"
)

type RemoteInstanceName struct {
	value string
}

func NewRemoteInstanceName(address RemoteInstanceAddress) (RemoteInstanceName, error) {
	u, err := url.Parse(address.String())
	if err != nil {
		return RemoteInstanceName{}, errors.Wrap(err, "could not parse the address")
	}

	return RemoteInstanceName{value: u.Host + u.Path}, nil
}

func (n RemoteInstanceName) String() string {
	return n.value
}

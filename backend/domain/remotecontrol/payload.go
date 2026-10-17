package remotecontrol

import "github.com/boreq/errors"

type Payload struct {
	value []byte
}

func NewPayload(b []byte) (Payload, error) {
	if len(b) == 0 {
		return Payload{}, errors.New("payload must not be empty")
	}
	return Payload{value: b}, nil
}

func (p Payload) Bytes() []byte {
	return p.value
}

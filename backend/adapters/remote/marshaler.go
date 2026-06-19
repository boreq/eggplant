package remote

import (
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type RemotePairingTokenSetPayload struct {
	RemoteInstanceID string `json:"remoteInstanceId"`
}

func MarshalRemotePairingTokenSet(event remotedomain.RemotePairingTokenSet) (*message.Message, error) {
	payload, err := json.Marshal(RemotePairingTokenSetPayload{
		RemoteInstanceID: event.RemoteInstanceID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal the payload")
	}

	return message.NewMessage(watermill.NewUUID(), payload), nil
}

func UnmarshalRemotePairingTokenSet(msg *message.Message) (remotedomain.RemotePairingTokenSet, error) {
	var payload RemotePairingTokenSetPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return remotedomain.RemotePairingTokenSet{}, errors.Wrap(err, "could not unmarshal the payload")
	}

	id, err := remotedomain.NewRemoteInstanceIDFromString(payload.RemoteInstanceID)
	if err != nil {
		return remotedomain.RemotePairingTokenSet{}, errors.Wrap(err, "invalid remote instance id")
	}

	return remotedomain.RemotePairingTokenSet{RemoteInstanceID: id}, nil
}

type RemotePairedPayload struct {
	RemoteInstanceID string `json:"remoteInstanceId"`
}

func MarshalRemotePaired(event remotedomain.RemotePaired) (*message.Message, error) {
	payload, err := json.Marshal(RemotePairedPayload{
		RemoteInstanceID: event.RemoteInstanceID.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal the payload")
	}

	return message.NewMessage(watermill.NewUUID(), payload), nil
}

func UnmarshalRemotePaired(msg *message.Message) (remotedomain.RemotePaired, error) {
	var payload RemotePairedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return remotedomain.RemotePaired{}, errors.Wrap(err, "could not unmarshal the payload")
	}

	id, err := remotedomain.NewRemoteInstanceIDFromString(payload.RemoteInstanceID)
	if err != nil {
		return remotedomain.RemotePaired{}, errors.Wrap(err, "invalid remote instance id")
	}

	return remotedomain.RemotePaired{RemoteInstanceID: id}, nil
}

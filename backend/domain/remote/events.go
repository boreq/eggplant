package remote

const (
	RemotePairingTokenSetEventName = "RemoteInstance.RemotePairingTokenSet.v1"
	RemotePairedEventName          = "RemoteInstance.RemotePaired.v1"
)

type Event interface {
	EventName() string
}

type RemotePairingTokenSet struct {
	RemoteInstanceID RemoteInstanceID
}

func (RemotePairingTokenSet) EventName() string {
	return RemotePairingTokenSetEventName
}

type RemotePaired struct {
	RemoteInstanceID RemoteInstanceID
}

func (RemotePaired) EventName() string {
	return RemotePairedEventName
}

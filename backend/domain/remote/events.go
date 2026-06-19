package remote

type Event interface {
	EventName() string
}

type RemotePairingTokenSet struct {
	RemoteInstanceID RemoteInstanceID
}

func (RemotePairingTokenSet) EventName() string {
	return "RemoteInstance.RemotePairingTokenSet.v1"
}

type RemotePaired struct {
	RemoteInstanceID RemoteInstanceID
}

func (RemotePaired) EventName() string {
	return "RemoteInstance.RemotePaired.v1"
}

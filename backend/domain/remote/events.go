package remote

const (
	RemotePairingTokenSetEventName = "RemoteInstance.RemotePairingTokenSet.v1"
	RemotePairedEventName          = "RemoteInstance.RemotePaired.v1"
)

type Event interface {
	EventName() string
}

type RemotePairingTokenSet struct {
	remoteInstanceID RemoteInstanceID
}

func NewRemotePairingTokenSet(remoteInstanceID RemoteInstanceID) RemotePairingTokenSet {
	return RemotePairingTokenSet{remoteInstanceID: remoteInstanceID}
}

func (e RemotePairingTokenSet) RemoteInstanceID() RemoteInstanceID {
	return e.remoteInstanceID
}

func (RemotePairingTokenSet) EventName() string {
	return RemotePairingTokenSetEventName
}

type RemotePaired struct {
	remoteInstanceID RemoteInstanceID
}

func NewRemotePaired(remoteInstanceID RemoteInstanceID) RemotePaired {
	return RemotePaired{remoteInstanceID: remoteInstanceID}
}

func (e RemotePaired) RemoteInstanceID() RemoteInstanceID {
	return e.remoteInstanceID
}

func (RemotePaired) EventName() string {
	return RemotePairedEventName
}

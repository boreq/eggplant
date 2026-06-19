package remote

type RemoteInstanceStatus struct {
	s string
}

var (
	RemoteInstanceStatusPairing = RemoteInstanceStatus{"PAIRING"}
	RemoteInstanceStatusHealthy = RemoteInstanceStatus{"HEALTHY"}
	RemoteInstanceStatusDead    = RemoteInstanceStatus{"DEAD"}
)
